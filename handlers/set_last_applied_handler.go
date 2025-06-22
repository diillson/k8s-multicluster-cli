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

type SetLastAppliedResult struct {
	Cluster executor.ClusterInfo
	Success bool
	Err     error
}

type SetLastAppliedHandler struct {
	Resource   string
	ApiGroup   string
	ApiVersion string
	Namespace  string
	Name       string
	Manifest   []byte
	KubeConfig string
	Results    []SetLastAppliedResult
	mu         sync.Mutex
}

func (h *SetLastAppliedHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, false, err)
		return err
	}
	gvr := schema.GroupVersionResource{Group: h.ApiGroup, Version: h.ApiVersion, Resource: h.Resource}
	var resourceIfc interface{}
	if h.Namespace == "" {
		resourceIfc = dynamicCli.Resource(gvr).Namespace("")
	} else {
		resourceIfc = dynamicCli.Resource(gvr).Namespace(h.Namespace)
	}
	ri, _ := resourceIfc.(interface {
		Get(context.Context, string, metav1.GetOptions) (*unstructured.Unstructured, error)
		Update(context.Context, *unstructured.Unstructured, metav1.UpdateOptions) (*unstructured.Unstructured, error)
	})
	obj, err := ri.Get(ctx, h.Name, metav1.GetOptions{})
	if err != nil {
		h.appendResult(cluster, false, err)
		return err
	}
	anns := obj.GetAnnotations()
	if anns == nil {
		anns = make(map[string]string)
	}
	anns["kubectl.kubernetes.io/last-applied-configuration"] = string(h.Manifest)
	obj.SetAnnotations(anns)
	_, err = ri.Update(ctx, obj, metav1.UpdateOptions{})
	h.appendResult(cluster, err == nil, err)
	return err
}
func (h *SetLastAppliedHandler) appendResult(cluster executor.ClusterInfo, success bool, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, SetLastAppliedResult{Cluster: cluster, Success: success, Err: err})
}
func (h *SetLastAppliedHandler) Result() interface{} { return h.Results }
