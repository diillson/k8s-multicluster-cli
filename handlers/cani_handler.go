package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
)

type CanIResult struct {
	Cluster   executor.ClusterInfo
	Verb      string
	Resource  string
	Name      string
	Namespace string
	Allowed   bool
	Reason    string
	Err       error
}

type CanIHandler struct {
	Verb       string
	Resource   string
	Name       string
	Namespace  string
	KubeConfig string
	Results    []CanIResult
	mu         sync.Mutex
}

func (h *CanIHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	clientset, _, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, false, "", err)
		return err
	}
	sar := &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: h.Namespace,
				Verb:      h.Verb,
				Resource:  h.Resource,
				Name:      h.Name,
			},
		},
	}
	resp, err := clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		h.appendResult(cluster, false, "", err)
		return err
	}
	h.appendResult(cluster, resp.Status.Allowed, resp.Status.Reason, nil)
	return nil
}

func (h *CanIHandler) appendResult(cluster executor.ClusterInfo, allowed bool, reason string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, CanIResult{
		Cluster: cluster, Verb: h.Verb, Resource: h.Resource,
		Name: h.Name, Namespace: h.Namespace,
		Allowed: allowed, Reason: reason, Err: err,
	})
}

func (h *CanIHandler) Result() interface{} { return h.Results }
