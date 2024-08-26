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

func ListConfigMaps(clientset *kubernetes.Clientset, clusterName string, namespaces []string) {
	var allConfigMaps []v1.ConfigMap

	for _, namespace := range namespaces {
		configMaps, err := clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.Errorf("Failed to list configmaps in cluster %s: %v", namespace, clusterName, err)
			continue
		}
		allConfigMaps = append(allConfigMaps, configMaps.Items...)
	}
	utils.PrintConfigMapTable(clusterName, allConfigMaps)

	fmt.Println("---")
}

func GetConfigMap(clientset *kubernetes.Clientset, clusterName string, namespaces []string, name string) {
	for _, namespace := range namespaces {
		configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			logrus.Errorf("Failed to get configmap %s in cluster %s: %v", name, namespace, clusterName, err)
			continue
		}

		fmt.Printf("Cluster: %s\nNamespace: %s\nConfigMap: %s\n\n", clusterName, namespace, name)
		for key, value := range configMap.Data {
			fmt.Printf("%s: %s\n", key, value)
		}
		fmt.Println("---")
	}
}
