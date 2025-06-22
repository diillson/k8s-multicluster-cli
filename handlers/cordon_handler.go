package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
)

type CordonResult struct {
	Cluster executor.ClusterInfo
	Node    string
	Success bool
	Message string
	Err     error
}

type CordonHandler struct {
	Node       string
	Unschedule bool // true=cordon, false=uncordon
	KubeConfig string
	Results    []CordonResult
	mu         sync.Mutex
}

func (h *CordonHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	clientset, _, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, h.Node, false, "", err)
		return err
	}
	node, err := clientset.CoreV1().Nodes().Get(ctx, h.Node, metav1.GetOptions{})
	if err != nil {
		h.appendResult(cluster, h.Node, false, "Node não encontrado", err)
		return err
	}
	origSched := node.Spec.Unschedulable
	node.Spec.Unschedulable = h.Unschedule
	if origSched == h.Unschedule {
		msg := "Node já estava no estado solicitado"
		h.appendResult(cluster, h.Node, true, msg, nil)
		return nil
	}
	_, err = clientset.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		h.appendResult(cluster, h.Node, false, "Falha ao atualizar node", err)
		return err
	}
	msg := "Node atualizado com sucesso"
	h.appendResult(cluster, h.Node, true, msg, nil)
	return nil
}

func (h *CordonHandler) appendResult(cluster executor.ClusterInfo, node string, success bool, msg string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, CordonResult{
		Cluster: cluster, Node: node, Success: success, Message: msg, Err: err,
	})
}
func (h *CordonHandler) Result() interface{} { return h.Results }
