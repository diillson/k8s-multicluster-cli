package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"log"
	"multicluster/config"
	"multicluster/k8s"
	"multicluster/utils"
	"sync"
)

var (
	podNamespaces string
	statusFilter  string
)

func init() {

	getPodsCmd.PersistentFlags().StringVarP(&configFile, "config", "c", utils.SetConfigDefault(), "Path to the clusters configuration file")
	getPodsCmd.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", utils.DefaultKubeconfigPath(), "Path to the kubeconfig file")
	getPodsCmd.PersistentFlags().StringVarP(&podNamespaces, "namespaces", "n", "", "Comma-separated list of namespaces to filter pods (default is all namespaces)")
	getPodsCmd.PersistentFlags().StringVarP(&statusFilter, "status", "s", "", "Status to filter pods (e.g. Running, Pending)")
	getPodsCmd.PersistentFlags().StringVarP(&clusterName, "cluster", "l", "", "Name of the cluster (if empty, get from all clusters)")

}

var getPodsCmd = &cobra.Command{
	Use:   "pods",
	Short: "Get pods from all clusters",
	Run: func(cmd *cobra.Command, args []string) {

		configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")

		namespaces := utils.GetNamespaces(podNamespaces)

		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		contexMap, err := config.GetContextMap(cfg)
		if err != nil {
			log.Fatalf("Failed to get context map: %v", err)
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
					k8s.ListPods(clientset, name, namespaces, statusFilter)
				}(name, contex)
			}
		}
		wg.Wait()
	},
}
