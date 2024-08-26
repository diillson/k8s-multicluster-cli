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
	deployNamespaces string
)

func init() {

	getDeploymentsCmd.PersistentFlags().StringVarP(&configFile, "config", "c", utils.SetConfigDefault(), "Path to the clusters configuration file")
	getDeploymentsCmd.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", utils.DefaultKubeconfigPath(), "Path to the kubeconfig file")
	getDeploymentsCmd.PersistentFlags().StringVarP(&deployNamespaces, "namespace", "n", "", "Namespace to filter deployments (default is all namespaces)")
}

var getDeploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "Get deployments from all clusters",
	Run: func(cmd *cobra.Command, args []string) {
		configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")

		namespaces := utils.GetNamespaces(deployNamespaces)

		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			logrus.Fatalf("Failed to load config: %v", err)
		}

		contexMap, err := config.GetContextMap(cfg)
		if err != nil {
			logrus.Fatalf("Failed to get context map: %v", err)
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

					k8s.ListDeployments(clientset, name, namespaces)
				}(name, contex)
			}
		}
		wg.Wait()
	},
}
