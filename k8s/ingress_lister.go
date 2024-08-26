package k8s

import (
	"context"
	"github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
	"multicluster/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListIngress(clientset *kubernetes.Clientset, clusterName string, namespaces []string) {
	var allingress []networkingv1.Ingress

	for _, namespace := range namespaces {
		ingresses, err := clientset.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.Errorf("Failed to list ingresses in cluster %s: %v", namespace, clusterName, err)
			continue
		}
		allingress = append(allingress, ingresses.Items...)
	}

	utils.PrintIngressTable(clusterName, allingress)
}
