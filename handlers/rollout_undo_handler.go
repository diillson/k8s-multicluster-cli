package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
	"strconv"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
)

type RolloutUndoResult struct {
	Cluster   executor.ClusterInfo
	Name      string
	Namespace string
	Revision  int64
	Success   bool
	Message   string
	Err       error
}
type RolloutUndoHandler struct {
	Resource   string
	ApiGroup   string
	ApiVersion string
	Namespace  string
	Name       string
	Revision   int64
	KubeConfig string
	Results    []RolloutUndoResult
	mu         sync.Mutex
}

func (h *RolloutUndoHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	clientset, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, h.Name, h.Namespace, h.Revision, false, "", err)
		return err
	}

	gvr := schema.GroupVersionResource{Group: h.ApiGroup, Version: h.ApiVersion, Resource: h.Resource}
	ns := h.Namespace
	if ns == "" {
		ns = "default"
	}
	objIfc := dynamicCli.Resource(gvr).Namespace(ns)
	obj, err := objIfc.Get(ctx, h.Name, metav1.GetOptions{})
	if err != nil {
		h.appendResult(cluster, h.Name, ns, h.Revision, false, fmt.Sprintf("Get error: %v", err), err)
		return err
	}

	kind := obj.GetKind()
	curRev := getCurrentRevision(obj)
	var tgtRev int64
	if h.Revision > 0 {
		tgtRev = h.Revision
	} else {
		if curRev > 1 {
			tgtRev = curRev - 1
		} else {
			h.appendResult(cluster, h.Name, ns, h.Revision, false, "Nenhuma revisão anterior disponível para undo", fmt.Errorf("no previous revision"))
			return fmt.Errorf("no previous revision")
		}
	}

	template, err := getTemplateForRevision(ctx, clientset, obj, kind, ns, h.Name, tgtRev)
	if err != nil {
		h.appendResult(cluster, h.Name, ns, tgtRev, false, err.Error(), err)
		return err
	}

	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": template,
		},
	}
	patchBytes, _ := json.Marshal(patch)
	_, err = objIfc.Patch(ctx, h.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		h.appendResult(cluster, h.Name, ns, tgtRev, false, fmt.Sprintf("Patch error: %v", err), err)
		return err
	}

	h.appendResult(cluster, h.Name, ns, tgtRev, true, "Undo realizado", nil)
	return nil
}

// ----- PRODUCTION UTILS ----- //

// Obtém o número da revisão atual do objeto.
func getCurrentRevision(obj *unstructured.Unstructured) int64 {
	ann, found, _ := unstructured.NestedString(obj.Object, "metadata", "annotations", "deployment.kubernetes.io/revision")
	if found && ann != "" {
		rev, err := strconv.ParseInt(ann, 10, 64)
		if err == nil {
			return rev
		}
	}
	// Para StatefulSet/DaemonSet (annotations diferentes)
	ann, found, _ = unstructured.NestedString(obj.Object, "metadata", "annotations", "controller.kubernetes.io/revision")
	if found && ann != "" {
		rev, err := strconv.ParseInt(ann, 10, 64)
		if err == nil {
			return rev
		}
	}
	return 1
}

// Busca a revisão/tpl em ReplicaSets (Deployment) ou ControllerRevision (StatefulSet/DaemonSet)
func getTemplateForRevision(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	obj *unstructured.Unstructured,
	kind, ns, name string,
	revision int64,
) (map[string]interface{}, error) {
	switch kind {
	case "Deployment":
		rsList, err := clientset.AppsV1().ReplicaSets(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("erro ao listar ReplicaSets: %v", err)
		}
		for _, rs := range rsList.Items {
			owner := false
			for _, o := range rs.OwnerReferences {
				if o.Kind == "Deployment" && o.Name == name && o.Controller != nil && *o.Controller {
					owner = true
					break
				}
			}
			if !owner {
				continue
			}
			rev, _ := strconv.ParseInt(rs.Annotations["deployment.kubernetes.io/revision"], 10, 64)
			if rev == revision {
				raw, _ := json.Marshal(rs.Spec.Template)
				var tpl map[string]interface{}
				json.Unmarshal(raw, &tpl)
				return tpl, nil
			}
		}
		return nil, fmt.Errorf("revisão %d não encontrada (ReplicaSets)", revision)
	case "StatefulSet", "DaemonSet":
		crList, err := clientset.AppsV1().ControllerRevisions(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("erro ao listar ControllerRevisions: %v", err)
		}
		for _, cr := range crList.Items {
			for _, o := range cr.OwnerReferences {
				if o.Kind == kind && o.Name == name && cr.Revision == revision {
					var tplMap map[string]interface{}
					// Aqui, cr.Data.Raw é um []byte em YAML/JSON
					if err := json.Unmarshal(cr.Data.Raw, &tplMap); err != nil {
						// Tenta fallback para YAML, se erro
						err = yaml.Unmarshal(cr.Data.Raw, &tplMap)
						if err != nil {
							return nil, fmt.Errorf("erro para parse ControllerRevision.Data: %v", err)
						}
					}
					return tplMap, nil
				}
			}
		}
		return nil, fmt.Errorf("revisão %d não encontrada (ControllerRevisions)", revision)
	default:
		return nil, fmt.Errorf("kind %s não suportado para undo", kind)
	}
}

func (h *RolloutUndoHandler) appendResult(cluster executor.ClusterInfo, name, ns string, revision int64, success bool, msg string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, RolloutUndoResult{
		Cluster: cluster, Name: name, Namespace: ns, Revision: revision, Success: success, Message: msg, Err: err,
	})
}
func (h *RolloutUndoHandler) Result() interface{} { return h.Results }
