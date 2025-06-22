package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sync"
)

type DescribeResult struct {
	Cluster   executor.ClusterInfo
	Namespace string
	Name      string
	Object    *unstructured.Unstructured
	Err       error
}

type DescribeResourceHandler struct {
	Resource   string
	ApiGroup   string
	ApiVersion string
	Namespaces []string
	Name       string
	KubeConfig string
	Results    []DescribeResult
	mu         sync.Mutex
}

func (h *DescribeResourceHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, "", h.Name, nil, err)
		return err
	}
	gvr := schema.GroupVersionResource{Group: h.ApiGroup, Version: h.ApiVersion, Resource: h.Resource}
	namespaces := h.Namespaces
	if len(namespaces) == 0 || (len(namespaces) == 1 && namespaces[0] == "") {
		namespaces = []string{""}
	}
	for _, ns := range namespaces {
		var resourceIfc dynamic.ResourceInterface
		if ns == "" {
			resourceIfc = dynamicCli.Resource(gvr).Namespace("")
		} else {
			resourceIfc = dynamicCli.Resource(gvr).Namespace(ns)
		}
		obj, err := resourceIfc.Get(ctx, h.Name, metav1.GetOptions{})
		h.appendResult(cluster, ns, h.Name, obj, err)
	}
	return nil
}

func (h *DescribeResourceHandler) appendResult(cluster executor.ClusterInfo, ns, name string, obj *unstructured.Unstructured, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, DescribeResult{
		Cluster: cluster, Namespace: ns, Name: name, Object: obj, Err: err,
	})
}

func (h *DescribeResourceHandler) Result() interface{} { return h.Results }
