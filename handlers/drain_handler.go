package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

type DrainResult struct {
	Cluster executor.ClusterInfo
	Node    string
	Success bool
	Message string
	Err     error
}

type DrainHandler struct {
	Node         string
	Force        bool
	IgnoreDaemon bool
	DeleteData   bool
	Timeout      int
	Namespace    string
	KubeConfig   string
	Results      []DrainResult
	mu           sync.Mutex
}

func (h *DrainHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	clientset, _, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, h.Node, false, "", err)
		return err
	}

	// 1. Cordone (set Unschedulable)
	node, err := clientset.CoreV1().Nodes().Get(ctx, h.Node, metav1.GetOptions{})
	if err != nil {
		h.appendResult(cluster, h.Node, false, "Node não encontrado", err)
		return err
	}
	if !node.Spec.Unschedulable {
		node.Spec.Unschedulable = true
		_, err = clientset.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
		if err != nil {
			h.appendResult(cluster, h.Node, false, "Falha ao cordon node", err)
			return err
		}
	}

	// 2. Lista todos pods agendados nesse node
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", h.Node),
	})
	if err != nil {
		h.appendResult(cluster, h.Node, false, "Falha ao listar pods", err)
		return err
	}

	// 3. Para cada pod, faz eviction (obedecendo flags)
	for i := range pods.Items {
		pod := &pods.Items[i]

		// Ignorar DaemonSet? (igual kubectl)
		if h.IgnoreDaemon && isDaemonSetPod(*pod) {
			continue
		}
		// Ignorar MirrorPod/StaticPod? (kubelet only)
		if isMirrorPod(*pod) {
			continue
		}
		// Se não for force, e pod não é evictável (usando sua regra), skip
		if !h.Force && !canBeEvicted(*pod) {
			h.appendResult(cluster, h.Node, false, fmt.Sprintf("Pod %s não pode ser removido", pod.Name), nil)
			return fmt.Errorf("pod %s não pode ser removido", pod.Name)
		}

		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			DeleteOptions: &metav1.DeleteOptions{
				GracePeriodSeconds: pointer.Int64Ptr(int64(h.Timeout)),
			},
		}

		err = clientset.PolicyV1().Evictions(pod.Namespace).Evict(ctx, eviction)
		if err != nil && !apierrors.IsNotFound(err) {
			h.appendResult(cluster, h.Node, false, fmt.Sprintf("Erro ao evict %s: %v", pod.Name, err), err)
			return err
		}
	}

	h.appendResult(cluster, h.Node, true, "Node drenado com sucesso", nil)
	return nil
}

// --- Helpers ---

func isDaemonSetPod(pod corev1.Pod) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "DaemonSet" && owner.Controller != nil && *owner.Controller {
			return true
		}
	}
	return false
}
func isMirrorPod(pod corev1.Pod) bool {
	_, ok := pod.Annotations["kubernetes.io/config.mirror"]
	return ok
}
func canBeEvicted(pod corev1.Pod) bool {
	return true
}

func (h *DrainHandler) appendResult(cluster executor.ClusterInfo, node string, success bool, msg string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, DrainResult{
		Cluster: cluster, Node: node, Success: success, Message: msg, Err: err,
	})
}
func (h *DrainHandler) Result() interface{} { return h.Results }
