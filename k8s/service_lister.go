package k8s

import (
	"context"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"multicluster/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListServices(clientset *kubernetes.Clientset, clusterName string, namespaces []string) {
	var allServices []v1.Service

	for _, namespace := range namespaces {
		services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.Errorf("Failed to list services in cluster %s: %v", clusterName, err)
			continue
		}
		allServices = append(allServices, services.Items...)
	}

	utils.PrintServiceTable(clusterName, allServices)
}
