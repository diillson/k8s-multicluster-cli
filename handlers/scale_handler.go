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
)

type ScaleResult struct {
	Cluster   executor.ClusterInfo
	Name      string
	Namespace string
	Replicas  int32
	Success   bool
	Message   string
	Err       error
}

type ScaleHandler struct {
	Resource   string
	ApiGroup   string
	ApiVersion string
	Namespace  string
	Name       string
	Replicas   int32
	KubeConfig string
	Results    []ScaleResult
	mu         sync.Mutex
}

func (h *ScaleHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, h.Name, h.Namespace, h.Replicas, false, "", err)
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
		Update(context.Context, *unstructured.Unstructured, metav1.UpdateOptions) (*unstructured.Unstructured, error)
	})
	obj, err := ri.Get(ctx, h.Name, metav1.GetOptions{})
	if err != nil {
		h.appendResult(cluster, h.Name, h.Namespace, h.Replicas, false, fmt.Sprintf("Get erro: %v", err), err)
		return err
	}
	if err := unstructured.SetNestedField(obj.Object, int64(h.Replicas), "spec", "replicas"); err != nil {
		h.appendResult(cluster, h.Name, h.Namespace, h.Replicas, false, "Falha set spec.replicas", err)
		return err
	}
	_, err = ri.Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		h.appendResult(cluster, h.Name, h.Namespace, h.Replicas, false, "Update erro", err)
		return err
	}
	h.appendResult(cluster, h.Name, h.Namespace, h.Replicas, true, "Scale OK", nil)
	return nil
}
func (h *ScaleHandler) appendResult(cluster executor.ClusterInfo, name, ns string, replicas int32, success bool, msg string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, ScaleResult{
		Cluster: cluster, Name: name, Namespace: ns, Replicas: replicas, Success: success, Message: msg, Err: err,
	})
}
func (h *ScaleHandler) Result() interface{} { return h.Results }
