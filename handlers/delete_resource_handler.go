package handlers

import (
	"context"
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sync"
)

type DeleteResult struct {
	Cluster executor.ClusterInfo
	Success bool
	Message string
	Err     error
}

type DeleteResourceHandler struct {
	Resource   string
	ApiGroup   string
	ApiVersion string
	Namespaces []string
	Name       string
	KubeConfig string
	Results    []DeleteResult
	mu         sync.Mutex
}

func (h *DeleteResourceHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, false, "", err)
		return err
	}

	gvr := schema.GroupVersionResource{
		Group:    h.ApiGroup,
		Version:  h.ApiVersion,
		Resource: h.Resource,
	}

	namespaces := h.Namespaces
	if len(namespaces) == 0 || (len(namespaces) == 1 && namespaces[0] == "") {
		namespaces = []string{""}
	}
	deleted := false
	for _, ns := range h.Namespaces {
		var resourceIfc dynamic.ResourceInterface
		if ns == "" {
			resourceIfc = dynamicCli.Resource(gvr).Namespace("") // cluster-scoped resource
		} else {
			resourceIfc = dynamicCli.Resource(gvr).Namespace(ns)
		}
		err := resourceIfc.Delete(ctx, h.Name, metav1.DeleteOptions{})
		if err != nil {
			h.appendResult(cluster, false, fmt.Sprintf("Erro: %v", err), err)
		} else {
			h.appendResult(cluster, true, "Deleted", nil)
			deleted = true
		}
	}
	if !deleted {
		return fmt.Errorf("delete not successful on any namespace")
	}
	return nil
}

func (h *DeleteResourceHandler) appendResult(cluster executor.ClusterInfo, success bool, message string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, DeleteResult{Cluster: cluster, Success: success, Message: message, Err: err})
}

func (h *DeleteResourceHandler) Result() interface{} {
	return h.Results
}
