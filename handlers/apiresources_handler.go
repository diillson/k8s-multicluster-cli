package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	"sync"
)

type APIResourcesResult struct {
	Cluster   executor.ClusterInfo
	Resources []string // ou []*metav1.APIResource para detalhamento
	Err       error
}

type APIResourcesHandler struct {
	KubeConfig string
	Results    []APIResourcesResult
	mu         sync.Mutex
}

func (h *APIResourcesHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	clientset, _, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, nil, err)
		return err
	}
	apiList, err := clientset.Discovery().ServerPreferredResources()
	if err != nil {
		h.appendResult(cluster, nil, err)
		return err
	}
	var resources []string
	for _, api := range apiList {
		for _, res := range api.APIResources {
			resources = append(resources, res.Name)
		}
	}
	h.appendResult(cluster, resources, nil)
	return nil
}

func (h *APIResourcesHandler) appendResult(cluster executor.ClusterInfo, resources []string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, APIResourcesResult{Cluster: cluster, Resources: resources, Err: err})
}

func (h *APIResourcesHandler) Result() interface{} { return h.Results }
