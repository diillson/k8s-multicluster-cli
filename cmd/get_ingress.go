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
	ingressNamespaces string
)

func init() {

	getIngressCmd.PersistentFlags().StringVarP(&configFile, "config", "c", utils.SetConfigDefault(), "Path to the clusters configuration file")
	getIngressCmd.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", utils.DefaultKubeconfigPath(), "Path to the kubeconfig file")
	getIngressCmd.PersistentFlags().StringVarP(&ingressNamespaces, "namespace", "n", "", "Namespace to filter ingresses (default is all namespaces)")
}

var getIngressCmd = &cobra.Command{
	Use:   "ingress",
	Short: "Get ingresses from all clusters",
	Run: func(cmd *cobra.Command, args []string) {
		configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")

		namespaces := utils.GetNamespaces(ingressNamespaces)

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

					k8s.ListIngress(clientset, name, namespaces)
				}(name, contex)
			}
		}
		wg.Wait()
	},
}
