package k8s

import (
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
)

func expandPath(path string) (string, error) {
	if len(path) > 2 && path[:2] == "~/" {
		homeDir := homedir.HomeDir()
		return filepath.Join(homeDir, path[2:]), nil
	}
	return path, nil
}

func CreateK8sClientFromContext(contextName string, kubeconfigPath string) (*kubernetes.Clientset, dynamic.Interface, error) {
	if kubeconfigPath == "" {
		utils.DefaultKubeconfigPath()
	}

	kubeconfigPath, err := expandPath(kubeconfigPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to expand kubeconfig path: %v", err)
	}

	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		configOverrides,
	).ClientConfig()
	if err != nil {
		return nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	return clientset, dynamicClient, nil
}

//func CreateK8sClient(cluster models.ClusterConfig, kubeconfigPath string) (*kubernetes.Clientset, dynamic.Interface, error) {
//	if kubeconfigPath == "" {
//		kubeconfigPath = "~/.kube/config"
//	}
//
//	kubeconfigPath, err := expandPath(kubeconfigPath)
//	if err != nil {
//		return nil, nil, fmt.Errorf("failed to expand kubeconfig path: %v", err)
//	}
//
//	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	clientset, err := kubernetes.NewForConfig(config)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	dynamicClient, err := dynamic.NewForConfig(config)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	return clientset, dynamicClient, nil
//}
