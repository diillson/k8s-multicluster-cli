package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type EditGetHandler struct {
	Resource   string
	ApiGroup   string
	ApiVersion string
	Namespace  string
	Name       string
	KubeConfig string
	ResultObj  *unstructured.Unstructured
}

func (h *EditGetHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	_, dynamicCli, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		return err
	}
	gvr := schema.GroupVersionResource{Group: h.ApiGroup, Version: h.ApiVersion, Resource: h.Resource}
	var ri interface{}
	if h.Namespace == "" {
		ri = dynamicCli.Resource(gvr).Namespace("")
	} else {
		ri = dynamicCli.Resource(gvr).Namespace(h.Namespace)
	}
	getIfc, _ := ri.(interface {
		Get(context.Context, string, metav1.GetOptions) (*unstructured.Unstructured, error)
	})
	obj, err := getIfc.Get(ctx, h.Name, metav1.GetOptions{})
	if err == nil {
		h.ResultObj = obj
	}
	return err
}

func (h *EditGetHandler) Result() interface{} { return h.ResultObj }
