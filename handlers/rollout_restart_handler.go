package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sync"
	"time"

	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
)

type RolloutRestartResult struct {
	Cluster   executor.ClusterInfo
	Name      string
	Namespace string
	Success   bool
	Message   string
	Err       error
}

type RolloutRestartHandler struct {
	Resource   string
	ApiGroup   string
	ApiVersion string
	Namespace  string
	Name       string
	KubeConfig string
	Results    []RolloutRestartResult
	mu         sync.Mutex
}

func (h *RolloutRestartHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, h.Name, h.Namespace, false, "", err)
		return err
	}
	gvr := schema.GroupVersionResource{Group: h.ApiGroup, Version: h.ApiVersion, Resource: h.Resource}
	var ri interface{}
	if h.Namespace == "" {
		ri = dynamicCli.Resource(gvr).Namespace("")
	} else {
		ri = dynamicCli.Resource(gvr).Namespace(h.Namespace)
	}
	resourceIfc, _ := ri.(interface {
		Patch(context.Context, string, types.PatchType, []byte, metav1.PatchOptions) (*unstructured.Unstructured, error)
	})
	now := time.Now().Format(time.RFC3339)
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"kubectl.kubernetes.io/restartedAt": now,
					},
				},
			},
		},
	}
	patchBytes, _ := json.Marshal(patch)
	_, err = resourceIfc.Patch(ctx, h.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		h.appendResult(cluster, h.Name, h.Namespace, false, fmt.Sprintf("Patch erro: %v", err), err)
		return err
	}
	h.appendResult(cluster, h.Name, h.Namespace, true, "Rollout restart ok", nil)
	return nil
}
func (h *RolloutRestartHandler) appendResult(cluster executor.ClusterInfo, name, ns string, success bool, msg string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, RolloutRestartResult{
		Cluster: cluster, Name: name, Namespace: ns, Success: success, Message: msg, Err: err,
	})
}
func (h *RolloutRestartHandler) Result() interface{} { return h.Results }
