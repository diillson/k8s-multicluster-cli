package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"multicluster/models"
	"path/filepath"
)

func LoadConfig(filename string) (*models.Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config models.Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func GetContextMap(config *models.Config) (map[string]string, error) {
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	kubeconfig, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	contextMap := make(map[string]string)
	for _, cluster := range config.Clusters {
		if _, ok := kubeconfig.Contexts[cluster.Context]; ok {
			contextMap[cluster.Name] = cluster.Context
		} else {
			return nil, fmt.Errorf("context %s not found in kubeconfig", cluster.Context)
		}
	}

	return contextMap, nil
}
