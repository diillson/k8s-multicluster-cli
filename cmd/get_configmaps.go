package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"multicluster/config"
	"multicluster/k8s"
	"multicluster/utils"
	"sync"
)

var (
	configMapNamespaces string
	configMapName       string
)

func init() {

	getConfigMapsCmd.PersistentFlags().StringVarP(&configFile, "config", "c", utils.SetConfigDefault(), "Path to the clusters configuration file")
	getConfigMapsCmd.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", utils.DefaultKubeconfigPath(), "Path to the kubeconfig file")
	getConfigMapsCmd.PersistentFlags().StringVarP(&configMapNamespaces, "namespaces", "n", "", "Comma-separated list of namespaces to filter pods (default is all namespaces")
	getConfigMapsCmd.PersistentFlags().StringVarP(&configMapName, "name", "m", "", "Name of the configmap (optional)")
	getConfigMapsCmd.PersistentFlags().StringVarP(&clusterName, "cluster", "l", "", "Name of the cluster (if empty, get from all clusters)")
}

var getConfigMapsCmd = &cobra.Command{
	Use:   "configmaps",
	Short: "Get configmaps from all clusters",
	Run: func(cmd *cobra.Command, args []string) {
		configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")

		namespaces := utils.GetNamespaces(configMapNamespaces)

		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			logrus.Fatalf("Failed to load config: %v", err)
		}

		contextMap, err := config.GetContextMap(cfg)
		if err != nil {
			logrus.Fatalf("Failed to get context map from kubeconfig: %v", err)
		}

		var wg sync.WaitGroup
		for name, context := range contextMap {
			if clusterName == "" || name == clusterName {
				wg.Add(1)
				go func(name, context string) {
					defer wg.Done()
					clientset, _, err := k8s.CreateK8sClientFromContext(context, kubeconfigPath)
					if err != nil {
						logrus.Errorf("Failed to create client for cluster %s: %v", name, err)
						return
					}

					if configMapName == "" {
						k8s.ListConfigMaps(clientset, name, namespaces)
					} else {
						k8s.GetConfigMap(clientset, name, namespaces, configMapName)
					}
				}(name, context)
			}
		}
		wg.Wait()
	},
}
