package cmd

import (
	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sync"
)

var (
	serviceNamespaces string
)

func init() {

	getServicesCmd.PersistentFlags().StringVarP(&configFile, "config", "c", utils.SetConfigDefault(), "Path to the clusters configuration file")
	getServicesCmd.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", utils.DefaultKubeconfigPath(), "Path to the kubeconfig file")
	getServicesCmd.PersistentFlags().StringVarP(&serviceNamespaces, "namespace", "n", "", "Namespace to filter services (default is all namespaces)")
	getServicesCmd.PersistentFlags().StringVarP(&clusterName, "cluster", "l", "", "Name of the cluster (if empty, get from all clusters)")
}

var getServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Get services from all clusters",
	Run: func(cmd *cobra.Command, args []string) {
		configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")

		namespaces := utils.GetNamespaces(serviceNamespaces)

		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			logrus.Fatalf("Failed to load config: %v", err)
		}

		contexMap, err := config.GetContextMap(cfg)
		if err != nil {
			logrus.Fatalf("Failed to get contexts: %v", err)
		}

		var wg sync.WaitGroup
		for name, contex := range contexMap {
			if clusterName == "" || name == clusterName {
				wg.Add(1)
				go func(name, contex string) {
					defer wg.Done()
					clientset, _, err := k8s.CreateK8sClientFromContext(contex, kubeconfigPath)
					if err != nil {
						logrus.Errorf("Failed to create client for cluster %s: %v", name, err)
						return
					}
					k8s.ListServices(clientset, name, namespaces)
				}(name, contex)
			}
		}
		wg.Wait()
	},
}
