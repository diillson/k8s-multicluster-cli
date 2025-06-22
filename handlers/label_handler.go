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

type LabelResult struct {
	Cluster executor.ClusterInfo
	Success bool
	Message string
	Err     error
}

type LabelHandler struct {
	Resource   string
	ApiGroup   string
	ApiVersion string
	Namespace  string
	Name       string
	Labels     map[string]string
	KubeConfig string
	Results    []LabelResult
	mu         sync.Mutex
}

func (h *LabelHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, false, "", err)
		return err
	}
	gvr := schema.GroupVersionResource{Group: h.ApiGroup, Version: h.ApiVersion, Resource: h.Resource}
	var ri interface{}
	if h.Namespace == "" {
		ri = dynamicCli.Resource(gvr).Namespace("")
	} else {
		ri = dynamicCli.Resource(gvr).Namespace(h.Namespace)
	}
	resIfc, _ := ri.(interface {
		Get(context.Context, string, metav1.GetOptions) (*unstructured.Unstructured, error)
		Update(context.Context, *unstructured.Unstructured, metav1.UpdateOptions) (*unstructured.Unstructured, error)
	})
	obj, err := resIfc.Get(ctx, h.Name, metav1.GetOptions{})
	if err != nil {
		h.appendResult(cluster, false, fmt.Sprintf("Get erro: %v", err), err)
		return err
	}
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	for k, v := range h.Labels {
		labels[k] = v
	}
	obj.SetLabels(labels)
	_, err = resIfc.Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		h.appendResult(cluster, false, fmt.Sprintf("Update erro: %v", err), err)
		return err
	}
	h.appendResult(cluster, true, "Label atualizada", nil)
	return nil
}
func (h *LabelHandler) appendResult(cluster executor.ClusterInfo, success bool, msg string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, LabelResult{Cluster: cluster, Success: success, Message: msg, Err: err})
}
func (h *LabelHandler) Result() interface{} { return h.Results }
