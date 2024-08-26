package k8s

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func GetPodLogs(clientset *kubernetes.Clientset, clusterName, namespace, podName, container string, follow bool) {
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		logrus.Errorf("Failed to get pod %s in cluster %s: %v", podName, clusterName, err)
		return
	}

	if container == "" && len(pod.Spec.Containers) > 0 {
		container = pod.Spec.Containers[0].Name
	}

	podLogOpts := corev1.PodLogOptions{
		Container: container,
		Follow:    follow,
	}

	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		logrus.Errorf("Failed to get logs for pod %s in cluster %s: %v", podName, clusterName, err)
		return
	}
	defer podLogs.Close()

	utils.PrintPodLogs(clusterName, namespace, podName, podLogs, follow)
}
