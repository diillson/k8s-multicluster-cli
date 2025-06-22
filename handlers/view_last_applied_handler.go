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

type ViewLastAppliedResult struct {
	Cluster     executor.ClusterInfo
	LastApplied string
	Err         error
}

type ViewLastAppliedHandler struct {
	Resource   string
	ApiGroup   string
	ApiVersion string
	Namespace  string
	Name       string
	KubeConfig string
	Results    []ViewLastAppliedResult
	mu         sync.Mutex
}

func (h *ViewLastAppliedHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, "", err)
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
	})
	obj, err := ri.Get(ctx, h.Name, metav1.GetOptions{})
	if err != nil {
		h.appendResult(cluster, "", err)
		return err
	}
	anns := obj.GetAnnotations()
	last := ""
	if anns != nil {
		last = anns["kubectl.kubernetes.io/last-applied-configuration"]
	}
	h.appendResult(cluster, last, nil)
	return nil
}
func (h *ViewLastAppliedHandler) appendResult(cluster executor.ClusterInfo, last string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, ViewLastAppliedResult{Cluster: cluster, LastApplied: last, Err: err})
}
func (h *ViewLastAppliedHandler) Result() interface{} { return h.Results }
