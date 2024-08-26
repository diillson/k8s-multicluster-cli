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
	secretNamespaces string
	secretName       string
)

func init() {

	getSecretsCmd.PersistentFlags().StringVarP(&configFile, "config", "c", utils.SetConfigDefault(), "Path to the clusters configuration file")
	getSecretsCmd.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", utils.DefaultKubeconfigPath(), "Path to the kubeconfig file")
	getSecretsCmd.PersistentFlags().StringVarP(&secretNamespaces, "namespace", "n", "", "Namespace to filter secrets (default is all namespaces)")
	getSecretsCmd.PersistentFlags().StringVarP(&secretName, "name", "s", "", "Name of the secret (optional)")
	getSecretsCmd.PersistentFlags().StringVarP(&clusterName, "cluster", "l", "", "Name of the cluster (if empty, get from all clusters)")
}

var getSecretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Get secrets from all clusters",
	Run: func(cmd *cobra.Command, args []string) {
		configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")

		namespaces := utils.GetNamespaces(secretNamespaces)

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
					if secretName == "" {
						k8s.ListSecrets(clientset, name, namespaces)
					} else {
						k8s.GetSecret(clientset, name, namespaces, secretName)
					}
				}(name, contex)
			}
		}
		wg.Wait()
	},
}
