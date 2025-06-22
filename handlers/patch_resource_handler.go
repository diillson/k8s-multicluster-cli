package handlers

import (
	"context"
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	typespkgs "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"strings"
	"sync"
)

type PatchResult struct {
	Cluster   executor.ClusterInfo
	Namespace string
	Name      string
	Success   bool
	Err       error
}

type PatchResourceHandler struct {
	Resource   string
	ApiGroup   string
	ApiVersion string
	Namespaces []string
	Name       string
	PatchType  string // "strategic", "merge", "json"
	Patch      []byte
	KubeConfig string
	Results    []PatchResult
	mu         sync.Mutex
}

func (h *PatchResourceHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, "", h.Name, false, err)
		return err
	}
	gvr := schema.GroupVersionResource{Group: h.ApiGroup, Version: h.ApiVersion, Resource: h.Resource}
	namespaces := h.Namespaces
	if len(namespaces) == 0 || (len(namespaces) == 1 && namespaces[0] == "") {
		namespaces = []string{""}
	}
	var patchType typespkgs.PatchType = typespkgs.StrategicMergePatchType
	switch strings.ToLower(h.PatchType) {
	case "strategic":
		patchType = typespkgs.StrategicMergePatchType
	case "merge":
		patchType = typespkgs.MergePatchType
	case "json":
		patchType = typespkgs.JSONPatchType
	}
	patched := false
	for _, ns := range namespaces {
		var resourceIfc dynamic.ResourceInterface
		if ns == "" {
			resourceIfc = dynamicCli.Resource(gvr).Namespace("") // cluster-scoped
		} else {
			resourceIfc = dynamicCli.Resource(gvr).Namespace(ns)
		}
		_, err := resourceIfc.Patch(ctx, h.Name, patchType, h.Patch, metav1.PatchOptions{})
		if err != nil {
			h.appendResult(cluster, ns, h.Name, false, err)
		} else {
			h.appendResult(cluster, ns, h.Name, true, nil)
			patched = true
		}
	}
	if !patched {
		return fmt.Errorf("patch not successful on any namespace")
	}
	return nil
}

func (h *PatchResourceHandler) appendResult(cluster executor.ClusterInfo, ns, name string, success bool, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, PatchResult{
		Cluster: cluster, Namespace: ns, Name: name, Success: success, Err: err,
	})
}

func (h *PatchResourceHandler) Result() interface{} { return h.Results }
