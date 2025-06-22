package handlers

import (
	"bytes"
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"

	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
)

type LogsJobResult struct {
	Cluster executor.ClusterInfo
	PodName string
	Stdout  string
	Err     error
}

type LogsJobHandler struct {
	Namespace  string
	JobName    string
	Container  string
	KubeConfig string
	Results    []LogsJobResult
	mu         sync.Mutex
}

func (h *LogsJobHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	clientset, _, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, "", "", err)
		return err
	}
	ns := h.Namespace
	if ns == "" {
		ns = "default"
	}
	// Lista todos pods controlados pelo Job
	pods, err := clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
		LabelSelector: "job-name=" + h.JobName,
	})
	if err != nil || len(pods.Items) == 0 {
		h.appendResult(cluster, "", "", fmt.Errorf("Job/pods não encontrados"))
		return err
	}
	for _, pod := range pods.Items {
		var buf bytes.Buffer
		req := clientset.CoreV1().Pods(ns).GetLogs(pod.Name, &corev1.PodLogOptions{
			Container: h.Container,
		})
		pl, err := req.Stream(ctx)
		if err != nil {
			h.appendResult(cluster, pod.Name, "", err)
			continue
		}
		buf.ReadFrom(pl)
		h.appendResult(cluster, pod.Name, buf.String(), nil)
		pl.Close()
	}
	return nil
}

func (h *LogsJobHandler) appendResult(cluster executor.ClusterInfo, podName, stdout string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, LogsJobResult{
		Cluster: cluster, PodName: podName, Stdout: stdout, Err: err,
	})
}
func (h *LogsJobHandler) Result() interface{} { return h.Results }
