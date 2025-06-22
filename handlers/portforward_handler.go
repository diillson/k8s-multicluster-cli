package handlers

import (
	"context"
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"strings"
	"sync"
)

// Resultado de cada port-forward, uso básico
type PortForwardResult struct {
	Cluster   executor.ClusterInfo
	Address   string
	Ports     []string
	Resource  string // pod ou service
	Name      string
	Namespace string
	Success   bool
	Err       error
}

type PortForwardHandler struct {
	Resource   string // "pod" or "service"
	Name       string
	Namespace  string
	Address    string
	Ports      []string
	KubeConfig string
	Results    []PortForwardResult
	mu         sync.Mutex
}

func (h *PortForwardHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	clientset, _, restConfig, err := k8s.CreateK8sClientsWithRestConfig(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, h.Address, h.Ports, h.Resource, h.Name, h.Namespace, false, err)
		return err
	}
	if h.Resource == "service" {
		// Descubra o pod por trás do service (por labelSelector), simplificadamente
		svc, err := clientset.CoreV1().Services(h.Namespace).Get(ctx, h.Name, metav1.GetOptions{})
		if err != nil {
			h.appendResult(cluster, h.Address, h.Ports, h.Resource, h.Name, h.Namespace, false, err)
			return err
		}
		selector := svc.Spec.Selector
		if len(selector) == 0 {
			h.appendResult(cluster, h.Address, h.Ports, h.Resource, h.Name, h.Namespace, false, fmt.Errorf("service tem selector vazio"))
			return fmt.Errorf("service tem selector vazio")
		}
		sel := make([]string, 0, len(selector))
		for k, v := range selector {
			sel = append(sel, fmt.Sprintf("%s=%s", k, v))
		}
		pods, err := clientset.CoreV1().Pods(h.Namespace).List(ctx, metav1.ListOptions{LabelSelector: strings.Join(sel, ",")})
		if err != nil || len(pods.Items) == 0 {
			h.appendResult(cluster, h.Address, h.Ports, h.Resource, h.Name, h.Namespace, false, fmt.Errorf("nenhum pod encontrado pro selector do svc"))
			return fmt.Errorf("nenhum pod encontrado")
		}
		// Faz port-forward para o primeiro pod do service
		return h.doPortForward(ctx, cluster, pods.Items[0].Name, restConfig)
	}
	// Se for pod:
	return h.doPortForward(ctx, cluster, h.Name, restConfig)
}

func (h *PortForwardHandler) doPortForward(ctx context.Context, cluster executor.ClusterInfo, pod string, restConfig *rest.Config) error {
	// Use o utilitário oficial de port-forward do client-go
	err := k8s.PortForward(ctx, restConfig, h.Namespace, pod, h.Address, h.Ports)
	h.appendResult(cluster, h.Address, h.Ports, h.Resource, pod, h.Namespace, err == nil, err)
	return err
}

func (h *PortForwardHandler) appendResult(cluster executor.ClusterInfo, address string, ports []string, resource, name, namespace string, success bool, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, PortForwardResult{
		Cluster: cluster, Address: address, Ports: ports, Resource: resource, Name: name, Namespace: namespace, Success: success, Err: err,
	})
}

func (h *PortForwardHandler) Result() interface{} { return h.Results }
