package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	"sync"
)

type APIVersionsResult struct {
	Cluster  executor.ClusterInfo
	Versions []string
	Err      error
}

type APIVersionsHandler struct {
	KubeConfig string
	Results    []APIVersionsResult
	mu         sync.Mutex
}

func (h *APIVersionsHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	clientset, _, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, nil, err)
		return err
	}
	groupList, err := clientset.Discovery().ServerGroups()
	if err != nil {
		h.appendResult(cluster, nil, err)
		return err
	}
	var versions []string
	for _, g := range groupList.Groups {
		for _, v := range g.Versions {
			versions = append(versions, v.GroupVersion)
		}
	}
	h.appendResult(cluster, versions, nil)
	return nil
}

func (h *APIVersionsHandler) appendResult(cluster executor.ClusterInfo, versions []string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, APIVersionsResult{Cluster: cluster, Versions: versions, Err: err})
}

func (h *APIVersionsHandler) Result() interface{} { return h.Results }
