package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
	"sync"
)

type RolloutHistoryResult struct {
	Cluster   executor.ClusterInfo
	Name      string
	Namespace string
	Revisions []string // YAML/json da revision ou breve resumo
	Err       error
}

type RolloutHistoryHandler struct {
	Resource   string
	ApiGroup   string
	ApiVersion string
	Namespace  string
	Name       string
	KubeConfig string
	Results    []RolloutHistoryResult
	mu         sync.Mutex
}

func (h *RolloutHistoryHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, h.Name, h.Namespace, nil, err)
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
	obj, err := ri.Get(ctx, h.Name, metav1.GetOptions{})
	if err != nil {
		h.appendResult(cluster, h.Name, h.Namespace, nil, err)
		return err
	}
	// Pegue as revisions do status, annotation, controller history
	revSummaries := extractRevisions(obj)
	h.appendResult(cluster, h.Name, h.Namespace, revSummaries, nil)
	return nil
}

func extractRevisions(obj *unstructured.Unstructured) []string {
	var revs []string
	// Para Deployment: lista de "controllerRevisions" geralmente vem como ReplicaSets (kubectl-style)
	if obj.GetKind() == "Deployment" {
		// Refs em status/replicaSets ou em annotation
		status, ok := obj.Object["status"].(map[string]interface{})
		if ok {
			if rs, found := status["replicaSets"].([]interface{}); found {
				for _, x := range rs {
					m, _ := x.(map[string]interface{})
					if rev, ok := m["name"].(string); ok {
						revs = append(revs, rev)
					}
				}
			}
		}
	}
	// Para StatefulSet: pode olhar annotations ou status/currentRevision/updateRevision
	if strings.HasSuffix(obj.GetKind(), "StatefulSet") {
		spec, ok := obj.Object["status"].(map[string]interface{})
		if ok {
			cr := ""
			if cr, ok := spec["currentRevision"].(string); ok && cr != "" {
				revs = append(revs, cr)
			}
			if ur, ok := spec["updateRevision"].(string); ok && ur != "" && ur != cr {
				revs = append(revs, ur)
			}
		}
	}
	return revs
}

func (h *RolloutHistoryHandler) appendResult(cluster executor.ClusterInfo, name, ns string, revs []string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, RolloutHistoryResult{
		Cluster: cluster, Name: name, Namespace: ns, Revisions: revs, Err: err,
	})
}
func (h *RolloutHistoryHandler) Result() interface{} { return h.Results }
