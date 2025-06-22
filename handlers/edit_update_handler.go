package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sync"
)

type EditUpdateResult struct {
	Cluster   executor.ClusterInfo
	Name      string
	Namespace string
	Success   bool
	Err       error
}
type EditUpdateHandler struct {
	Resource   string
	ApiGroup   string
	ApiVersion string
	Namespace  string
	Name       string
	NewObj     map[string]interface{}
	KubeConfig string
	Results    []EditUpdateResult
	mu         sync.Mutex
}

func (h *EditUpdateHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, h.Name, h.Namespace, false, err)
		return err
	}
	gvr := schema.GroupVersionResource{Group: h.ApiGroup, Version: h.ApiVersion, Resource: h.Resource}
	var ri interface{}
	if h.Namespace == "" {
		ri = dynamicCli.Resource(gvr).Namespace("")
	} else {
		ri = dynamicCli.Resource(gvr).Namespace(h.Namespace)
	}
	updateIfc, _ := ri.(interface {
		Update(context.Context, *unstructured.Unstructured, metav1.UpdateOptions) (*unstructured.Unstructured, error)
	})
	obj := &unstructured.Unstructured{Object: h.NewObj}
	_, err = updateIfc.Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		h.appendResult(cluster, h.Name, h.Namespace, false, err)
		return err
	}
	h.appendResult(cluster, h.Name, h.Namespace, true, nil)
	return nil
}
func (h *EditUpdateHandler) appendResult(cluster executor.ClusterInfo, name, ns string, success bool, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, EditUpdateResult{
		Cluster: cluster, Name: name, Namespace: ns, Success: success, Err: err,
	})
}
func (h *EditUpdateHandler) Result() interface{} { return h.Results }
