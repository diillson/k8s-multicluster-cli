package k8s

import (
	"context"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"multicluster/utils"
)

func ListPods(clientset *kubernetes.Clientset, clusterName string, namespaces []string, statusFilter string) {
	var allPods []v1.Pod

	for _, namespace := range namespaces {
		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.Errorf("Failed to list pods in cluster %s: %v", namespace, clusterName, err)
			continue
		}
		allPods = append(allPods, pods.Items...)
	}

	utils.PrintPodTable(clusterName, allPods, statusFilter)
}
