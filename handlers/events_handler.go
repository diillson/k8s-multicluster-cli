package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
)

type EventsResult struct {
	Cluster   executor.ClusterInfo
	Namespace string
	Events    []v1.Event
	Err       error
}

type EventsHandler struct {
	Namespaces []string
	KubeConfig string
	Results    []EventsResult
	mu         sync.Mutex
}

func (h *EventsHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	clientset, _, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, "", nil, err)
		return err
	}
	namespaces := h.Namespaces
	if len(namespaces) == 0 || (len(namespaces) == 1 && namespaces[0] == "") {
		namespaces = []string{""}
	}
	for _, ns := range namespaces {
		var events *v1.EventList
		var err error
		if ns != "" {
			events, err = clientset.CoreV1().Events(ns).List(ctx, metav1.ListOptions{})
		} else {
			// List all events em todos namespaces
			events, err = clientset.CoreV1().Events("").List(ctx, metav1.ListOptions{})
		}
		h.appendResult(cluster, ns, events.Items, err)
	}
	return nil
}

func (h *EventsHandler) appendResult(cluster executor.ClusterInfo, ns string, evs []v1.Event, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, EventsResult{Cluster: cluster, Namespace: ns, Events: evs, Err: err})
}

func (h *EventsHandler) Result() interface{} { return h.Results }
