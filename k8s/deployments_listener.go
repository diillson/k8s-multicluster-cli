package k8s

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListDeployments(clientset *kubernetes.Clientset, clusterName string, namespaces []string) {
	var allDeployments []appsv1.Deployment

	for _, namespace := range namespaces {
		deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.Errorf("Failed to list deployments in cluster %s: %v", clusterName, err)
			continue
		}
		allDeployments = append(allDeployments, deployments.Items...)
	}
	utils.PrintDeploymentTable(clusterName, allDeployments)
}
