package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/models"
	"strings"
	"sync"

	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type GetResourceHandler struct {
	Resource      string
	ApiGroup      string
	ApiVersion    string
	Namespaces    []string
	Names         []string
	LabelSelector string
	FieldSelector string
	KubeConfig    string
	OutputFormat  string
	Results       []models.GenericResourceListResult
	mu            sync.Mutex
}

func (h *GetResourceHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, nil, err)
		return err
	}

	gvr := schema.GroupVersionResource{
		Group:    h.ApiGroup,
		Version:  h.ApiVersion,
		Resource: h.Resource,
	}

	var allItems []unstructured.Unstructured
	if len(h.Names) == 0 {
		namespaces := h.Namespaces
		if len(namespaces) == 0 || (len(namespaces) == 1 && namespaces[0] == "") {
			namespaces = []string{""}
		}
		for _, ns := range namespaces {
			var ri interface{}
			if ns == "" {
				ri = dynamicCli.Resource(gvr)
			} else {
				ri = dynamicCli.Resource(gvr).Namespace(ns)
			}
			// Cast para DynamicResourceInterface
			r, _ := ri.(interface {
				List(context.Context, metav1.ListOptions) (*unstructured.UnstructuredList, error)
			})
			if r == nil {
				continue
			}
			list, err := r.List(ctx, metav1.ListOptions{
				LabelSelector: h.LabelSelector,
				FieldSelector: h.FieldSelector,
			})
			if err != nil {
				continue
			}
			allItems = append(allItems, list.Items...)
		}
	} else {
		for _, ns := range h.Namespaces {
			for _, name := range h.Names {
				var ri interface{}
				if ns == "" {
					ri = dynamicCli.Resource(gvr)
				} else {
					ri = dynamicCli.Resource(gvr).Namespace(ns)
				}
				r, _ := ri.(interface {
					Get(context.Context, string, metav1.GetOptions) (*unstructured.Unstructured, error)
				})
				if r == nil {
					continue
				}
				res, err := r.Get(ctx, name, metav1.GetOptions{})
				if err != nil {
					continue
				}
				allItems = append(allItems, *res)
			}
		}
	}

	h.appendResult(cluster, allItems, nil)
	return nil
}

func (h *GetResourceHandler) appendResult(cluster executor.ClusterInfo, items []unstructured.Unstructured, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, models.GenericResourceListResult{Cluster: cluster, Items: items, Err: err})
}

func (h *GetResourceHandler) Result() interface{} {
	return h.Results
}

func InferResourceFromInput(input string) (group, version, resource string) {
	resource = input
	group, version = "", "v1"
	if strings.Contains(input, ".") {
		parts := strings.SplitN(input, ".", 2)
		resource = parts[0]
		group = parts[1]
		version = "v1"
	}
	return group, version, resource
}
