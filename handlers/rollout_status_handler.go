package handlers

import (
	"context"
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sync"
	"time"
)

type RolloutStatusResult struct {
	Cluster   executor.ClusterInfo
	Name      string
	Namespace string
	Done      bool
	Message   string
	Err       error
}

type RolloutStatusHandler struct {
	Resource   string
	ApiGroup   string
	ApiVersion string
	Namespace  string
	Name       string
	Timeout    int
	KubeConfig string
	Results    []RolloutStatusResult
	mu         sync.Mutex
}

func (h *RolloutStatusHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, h.Name, h.Namespace, false, "", err)
		return err
	}
	gvr := schema.GroupVersionResource{Group: h.ApiGroup, Version: h.ApiVersion, Resource: h.Resource}
	var resourceIfc interface{}
	if h.Namespace == "" {
		resourceIfc = dynamicCli.Resource(gvr).Namespace("")
	} else {
		resourceIfc = dynamicCli.Resource(gvr).Namespace(h.Namespace)
	}
	ri, _ := resourceIfc.(interface {
		Get(context.Context, string, metav1.GetOptions) (*unstructured.Unstructured, error)
	})

	timeout := time.Now().Add(time.Duration(h.Timeout) * time.Second)
	for {
		obj, err := ri.Get(ctx, h.Name, metav1.GetOptions{})
		if err != nil {
			h.appendResult(cluster, h.Name, h.Namespace, false, fmt.Sprintf("Get erro: %v", err), err)
			return err
		}
		if rolloutComplete(obj) {
			h.appendResult(cluster, h.Name, h.Namespace, true, "Rollout completado", nil)
			return nil
		}
		if time.Now().After(timeout) {
			h.appendResult(cluster, h.Name, h.Namespace, false, "Timeout esperando rollout", fmt.Errorf("timeout"))
			return fmt.Errorf("timeout")
		}
		time.Sleep(2 * time.Second)
	}
}

func rolloutComplete(obj *unstructured.Unstructured) bool {
	// Heurística simples para resource "deployment":
	// deployment: status.updatedReplicas == spec.replicas && availableReplicas == spec.replicas
	specReplicas, _, _ := unstructured.NestedInt64(obj.Object, "spec", "replicas")
	updatedReplicas, _, _ := unstructured.NestedInt64(obj.Object, "status", "updatedReplicas")
	availableReplicas, _, _ := unstructured.NestedInt64(obj.Object, "status", "availableReplicas")
	generation, _, _ := unstructured.NestedInt64(obj.Object, "metadata", "generation")
	obsGen, _, _ := unstructured.NestedInt64(obj.Object, "status", "observedGeneration")
	return generation > 0 && obsGen == generation && specReplicas == updatedReplicas && specReplicas == availableReplicas && specReplicas > 0
}

func (h *RolloutStatusHandler) appendResult(cluster executor.ClusterInfo, name, ns string, done bool, msg string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, RolloutStatusResult{
		Cluster: cluster, Name: name, Namespace: ns, Done: done, Message: msg, Err: err,
	})
}
func (h *RolloutStatusHandler) Result() interface{} { return h.Results }
