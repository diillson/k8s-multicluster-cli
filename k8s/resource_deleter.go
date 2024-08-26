package k8s

import (
	"context"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DeleteResource(clientset *kubernetes.Clientset, clusterName, namespace, resourceType, resourceName string) {
	var err error
	switch resourceType {
	case "pod":
		err = clientset.CoreV1().Pods(namespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
	case "service":
		err = clientset.CoreV1().Services(namespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
	case "deployment":
		err = clientset.AppsV1().Deployments(namespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
	// Adicione outros tipos de recursos conforme necessário
	default:
		logrus.Errorf("Unsupported resource type: %s", resourceType)
		return
	}

	if err != nil {
		logrus.Errorf("Failed to delete %s %s in cluster %s: %v", resourceType, resourceName, clusterName, err)
	} else {
		logrus.Errorf("Deleted %s %s in cluster %s", resourceType, resourceName, clusterName)
	}
}
