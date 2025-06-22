package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"strings"
	"sync"
)

type ApplyResult struct {
	Cluster   executor.ClusterInfo
	Resource  string
	Namespace string
	Name      string
	Success   bool
	Err       error
}

type ApplyResourceHandler struct {
	ManifestFiles []string
	Namespace     string
	KubeConfig    string
	Results       []ApplyResult
	mu            sync.Mutex
}

func (h *ApplyResourceHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, "", "", "", false, err)
		return err
	}
	for _, file := range h.ManifestFiles {
		manifestData, err := ioutil.ReadFile(file)
		if err != nil {
			h.appendResult(cluster, file, "", "", false, err)
			continue
		}
		decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(manifestData)), 4096)
		for {
			var raw map[string]interface{}
			if err := decoder.Decode(&raw); err != nil {
				if err.Error() == "EOF" {
					break
				}
				h.appendResult(cluster, file, "", "", false, err)
				break
			}
			obj := &unstructured.Unstructured{Object: raw}
			gvk := obj.GroupVersionKind()
			gvr := schema.GroupVersionResource{
				Group: gvk.Group, Version: gvk.Version, Resource: strings.ToLower(gvk.Kind) + "s",
			}
			ns := obj.GetNamespace()
			if h.Namespace != "" {
				obj.SetNamespace(h.Namespace)
				ns = h.Namespace
			}
			ri := dynamicCli.Resource(gvr)
			// decide se usa namespace ou cluster-scoped
			var res *unstructured.Unstructured
			if ns != "" {
				res, err = ri.Namespace(ns).Create(ctx, obj, metav1.CreateOptions{})
			} else {
				res, err = ri.Create(ctx, obj, metav1.CreateOptions{})
			}
			if err != nil {
				h.appendResult(cluster, gvr.Resource, ns, obj.GetName(), false, err)
			} else {
				h.appendResult(cluster, gvr.Resource, ns, res.GetName(), true, nil)
			}
		}
	}
	return nil
}

func (h *ApplyResourceHandler) appendResult(cluster executor.ClusterInfo, resource, ns, name string, success bool, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, ApplyResult{
		Cluster:   cluster,
		Resource:  resource,
		Namespace: ns,
		Name:      name,
		Success:   success,
		Err:       err,
	})
}

func (h *ApplyResourceHandler) Result() interface{} {
	return h.Results
}
