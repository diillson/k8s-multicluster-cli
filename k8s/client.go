package k8s

import (
	"fmt"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

// Recebe contextName e caminho kubeconfig, retorna clientset e dynamicClient igual kubectl.
func CreateK8sClients(contextName, kubeconfigPath string) (*kubernetes.Clientset, dynamic.Interface, error) {
	if kubeconfigPath == "" {
		kubeconfigPath = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		configOverrides,
	).ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build k8s config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create k8s clientset: %w", err)
	}
	dynamicCli, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}
	return clientset, dynamicCli, nil
}

func CreateK8sClientsWithRestConfig(contextName, kubeconfigPath string) (*kubernetes.Clientset, dynamic.Interface, *rest.Config, error) {
	if kubeconfigPath == "" {
		kubeconfigPath = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		configOverrides,
	).ClientConfig()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to build k8s config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create k8s clientset: %w", err)
	}
	dynamicCli, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}
	return clientset, dynamicCli, config, nil
}
