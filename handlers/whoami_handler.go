package handlers

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
)

type WhoamiResult struct {
	Cluster  executor.ClusterInfo
	Username string
	UID      string
	Groups   []string
	Extras   map[string]authv1.ExtraValue
	Err      error
}

type WhoamiHandler struct {
	KubeConfig string
	Results    []WhoamiResult
	mu         sync.Mutex
}

func (h *WhoamiHandler) RunOnCluster(ctx context.Context, cluster executor.ClusterInfo, kubeConfigPath string) error {
	clientset, _, err := k8s.CreateK8sClients(cluster.Context, kubeConfigPath)
	if err != nil {
		h.appendResult(cluster, "", "", nil, nil, err)
		return err
	}
	ssr, err := clientset.AuthenticationV1().SelfSubjectReviews().Create(ctx, &authv1.SelfSubjectReview{}, metav1.CreateOptions{})
	if err != nil {
		h.appendResult(cluster, "", "", nil, nil, err)
		return err
	}
	u := ssr.Status.UserInfo
	h.appendResult(cluster, u.Username, u.UID, u.Groups, u.Extra, nil)
	return nil
}

func (h *WhoamiHandler) appendResult(cluster executor.ClusterInfo, username, uid string, groups []string, extras map[string]authv1.ExtraValue, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Results = append(h.Results, WhoamiResult{
		Cluster: cluster, Username: username, UID: uid, Groups: groups, Extras: extras, Err: err,
	})
}

func (h *WhoamiHandler) Result() interface{} { return h.Results }
