package handlers

import (
	"context"
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sync"
)

type AnnotateResult struct {
	Cluster executor.ClusterInfo
	Success bool
	Message string
	Err     error
}

type AnnotateHandler struct {
	Resource    string
	ApiGroup    string
	ApiVersion  string
	Namespace   string
	Name        string
	Annotations map[string]string
	KubeConfig  string
	Results     []AnnotateResult
	mu          sync.Mutex
}

func (h *AnnotateHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, false, "", err)
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
		h.appendResult(cluster, false, fmt.Sprintf("Get erro: %v", err), err)
		return err
	}

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	for k, v := range h.Annotations {
		annotations[k] = v
	}
	obj.SetAnnotations(annotations)

	_, err = ri.Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		h.appendResult(cluster, false, fmt.Sprintf("Update erro: %v", err), err)
		return err
	}
	h.appendResult(cluster, true, "Annotation definida com sucesso", nil)
	return nil
}

func (h *AnnotateHandler) appendResult(cluster executor.ClusterInfo, success bool, msg string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, AnnotateResult{
		Cluster: cluster,
		Success: success,
		Message: msg,
		Err:     err,
	})
}

func (h *AnnotateHandler) Result() interface{} {
	return h.Results
}
