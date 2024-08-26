package k8s

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DescribePod(clientset *kubernetes.Clientset, clusterName, namespace, podName string) error {
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting pod %s in namespace %s: %v", podName, namespace, err)
	}

	fmt.Printf("Cluster: %s\nNamespace: %s\nPod: %s\n\n", clusterName, namespace, podName)
	fmt.Printf("Status: %s\nNode: %s\nIP: %s\n\n", pod.Status.Phase, pod.Spec.NodeName, pod.Status.PodIP)
	fmt.Println("Containers:")
	for _, container := range pod.Spec.Containers {
		fmt.Printf("- Name: %s\n  Image: %s\n  Ports: %v\n  Environment:\n", container.Name, container.Image, container.Ports)
		for _, env := range container.Env {
			fmt.Printf("    %s: %s\n", env.Name, env.Value)
		}
	}
	fmt.Println("---")
	return nil
}
