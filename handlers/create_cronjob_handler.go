package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
	"sync"
)

type CreateCronJobResult struct {
	Cluster   executor.ClusterInfo
	Name      string
	Namespace string
	Success   bool
	Err       error
}

type CreateCronJobHandler struct {
	Namespace  string
	Manifest   []byte
	KubeConfig string
	Results    []CreateCronJobResult
	mu         sync.Mutex
}

func (h *CreateCronJobHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, "", "", false, err)
		return err
	}
	// Parse manifest para unstructured
	var objMap map[string]interface{}
	if err := yaml.Unmarshal(h.Manifest, &objMap); err != nil {
		h.appendResult(cluster, "", "", false, err)
		return err
	}
	u := &unstructured.Unstructured{Object: objMap}
	ns := h.Namespace
	if ns == "" {
		ns = u.GetNamespace() // fallback para namespace do objeto
	}
	if ns == "" {
		ns = "default"
		u.SetNamespace(ns)
	}
	gvr := schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "cronjobs"}
	ri := dynamicCli.Resource(gvr).Namespace(ns)
	_, err = ri.Create(ctx, u, metav1.CreateOptions{})
	if err != nil {
		h.appendResult(cluster, u.GetName(), ns, false, err)
		return err
	}
	h.appendResult(cluster, u.GetName(), ns, true, nil)
	return nil
}
func (h *CreateCronJobHandler) appendResult(cluster executor.ClusterInfo, name, ns string, success bool, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, CreateCronJobResult{
		Cluster: cluster, Name: name, Namespace: ns, Success: success, Err: err,
	})
}
func (h *CreateCronJobHandler) Result() interface{} { return h.Results }
