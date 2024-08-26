package k8s

import (
	"context"
	"github.com/sirupsen/logrus"
	"multicluster/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListNodes(clientset *kubernetes.Clientset, clusterName string) {
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("Failed to list nodes in cluster %s: %v", clusterName, err)
		return
	}

	utils.PrintNodeTable(clusterName, nodes.Items)
}
