package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	"io"
	corev1 "k8s.io/api/core/v1"
	"sync"
)

type LogsResult struct {
	Cluster   executor.ClusterInfo
	Namespace string
	PodName   string
	Container string
	Logs      string // log buffer (para tail), pode ser um []string ou buffer
	Err       error
}

type LogsHandler struct {
	PodName    string
	Namespaces []string
	Container  string
	Follow     bool
	Tail       int64
	Since      int64
	KubeConfig string

	Results []LogsResult
	mu      sync.Mutex
}

func (h *LogsHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	clientset, _, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, "", h.PodName, h.Container, "", err)
		return err
	}
	for _, ns := range h.Namespaces {
		podLogOpts := corev1.PodLogOptions{
			Container: h.Container,
			Follow:    h.Follow,
			TailLines: &h.Tail,
		}
		if h.Since > 0 {
			secs := int64(h.Since)
			podLogOpts.SinceSeconds = &secs
		}
		req := clientset.CoreV1().Pods(ns).GetLogs(h.PodName, &podLogOpts)
		podLogs, err := req.Stream(ctx)
		if err != nil {
			h.appendResult(cluster, ns, h.PodName, h.Container, "", err)
			continue
		}
		defer podLogs.Close()
		buf := make([]byte, 0, 1024*16)
		chunk := make([]byte, 4096)
		for {
			n, err := podLogs.Read(chunk)
			if n > 0 {
				buf = append(buf, chunk[:n]...)
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				h.appendResult(cluster, ns, h.PodName, h.Container, string(buf), err)
				break
			}
		}
		h.appendResult(cluster, ns, h.PodName, h.Container, string(buf), nil)
	}
	return nil
}

func (h *LogsHandler) appendResult(cluster executor.ClusterInfo, ns, pod, container, logs string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, LogsResult{
		Cluster: cluster, Namespace: ns, PodName: pod, Container: container, Logs: logs, Err: err,
	})
}

func (h *LogsHandler) Result() interface{} { return h.Results }
