package k8s

import (
	"context"
	"github.com/sirupsen/logrus"
	"multicluster/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListNamespaces(clientset *kubernetes.Clientset, clusterName string) {
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("Failed to list namespaces in cluster %s: %v", clusterName, err)
		return
	}

	utils.PrintNamespaceTable(clusterName, namespaces.Items)
}
