package k8s

import (
	"context"
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListSecrets(clientset *kubernetes.Clientset, clusterName string, namespaces []string) {
	var allSecrets []v1.Secret

	for _, namespace := range namespaces {
		secrets, err := clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.Errorf("Failed to list secrets in cluster %s: %v", clusterName, err)
			continue
		}
		allSecrets = append(allSecrets, secrets.Items...)
	}

	utils.PrintSecretTable(clusterName, allSecrets)
}

func GetSecret(clientset *kubernetes.Clientset, clusterName string, namespaces []string, secretName string) {
	for _, namespace := range namespaces {
		secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			logrus.Errorf("Failed to get secret %s in cluster %s: %v", secretName, namespace, clusterName, err)
			continue
		}

		fmt.Printf("Cluster: %s\nNamespace: %s\nSecret: %s\n\n", clusterName, namespace, secretName)
		for key, value := range secret.Data {
			fmt.Printf("%s: %s\n", key, value)
		}
		fmt.Println("")
		fmt.Println("")
		fmt.Println("")

	}
}
