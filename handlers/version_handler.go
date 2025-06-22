package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	"sync"
)

type VersionResult struct {
	Cluster     executor.ClusterInfo
	ServerMajor string
	ServerMinor string
	GitVersion  string
	Platform    string
	Err         error
}

type VersionHandler struct {
	KubeConfig string
	Results    []VersionResult
	mu         sync.Mutex
}

func (h *VersionHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	clientset, _, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, "", "", "", "", err)
		return err
	}
	info, err := clientset.Discovery().ServerVersion()
	if err != nil {
		h.appendResult(cluster, "", "", "", "", err)
		return err
	}
	h.appendResult(cluster, info.Major, info.Minor, info.GitVersion, info.Platform, nil)
	return nil
}

func (h *VersionHandler) appendResult(cluster executor.ClusterInfo, major, minor, git, platform string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, VersionResult{
		Cluster: cluster, ServerMajor: major, ServerMinor: minor, GitVersion: git, Platform: platform, Err: err,
	})
}

func (h *VersionHandler) Result() interface{} { return h.Results }
