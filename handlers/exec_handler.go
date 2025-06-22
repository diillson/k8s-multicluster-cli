package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	"sync"
)

type ExecResult struct {
	Cluster   executor.ClusterInfo
	Namespace string
	PodName   string
	Container string
	Command   []string
	Output    string // resultado stdout/stderr do comando
	Err       error
}

type ExecHandler struct {
	PodName    string
	Namespaces []string
	Container  string
	Command    []string
	Stdin      bool
	TTY        bool
	KubeConfig string

	Results []ExecResult
	mu      sync.Mutex
}

func (h *ExecHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	for _, ns := range h.Namespaces {
		clientset, _, restConfig, err := k8s.CreateK8sClientsWithRestConfig(cluster.Context, kubeConfigPath)
		if err != nil {
			h.appendResult(cluster, ns, h.PodName, h.Container, h.Command, "", err)
			continue
		}
		// Chama função canônica de exec em produção
		output, err := k8s.ExecPodToBuffer(ctx, clientset, restConfig, ns, h.PodName, h.Container, h.Command, h.Stdin, h.TTY)
		h.appendResult(cluster, ns, h.PodName, h.Container, h.Command, output, err)
	}
	return nil
}

func (h *ExecHandler) appendResult(cluster executor.ClusterInfo, ns, pod, container string, command []string, output string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, ExecResult{
		Cluster: cluster, Namespace: ns, PodName: pod, Container: container, Command: command, Output: output, Err: err})
}

func (h *ExecHandler) Result() interface{} { return h.Results }
