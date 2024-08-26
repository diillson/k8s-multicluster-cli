package cmd

import (
	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"sync"
)

var (
	deleteNamespace string
	resourceType    string
	resourceName    string
)

func init() {
	rootCmd.AddCommand(deleteResourceCmd)

	deleteResourceCmd.PersistentFlags().StringVarP(&configFile, "config", "c", utils.SetConfigDefault(), "Path to the clusters configuration file")
	deleteResourceCmd.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", utils.DefaultKubeconfigPath(), "Path to the kubeconfig file")
	deleteResourceCmd.PersistentFlags().StringVarP(&deleteNamespace, "namespace", "n", "", "Namespace of the resource")
	deleteResourceCmd.PersistentFlags().StringVarP(&resourceType, "type", "t", "", "Type of the resource (pod, service, deployment, etc.)")
	deleteResourceCmd.PersistentFlags().StringVarP(&resourceName, "name", "r", "", "Name of the resource")
	deleteResourceCmd.PersistentFlags().StringVarP(&clusterName, "cluster", "l", "", "Name of the cluster (if empty, delete from all clusters)")
}

var deleteResourceCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a resource from one or all clusters",
	Run: func(cmd *cobra.Command, args []string) {

		if !cmd.Flags().Changed("config") || os.Getenv("MC_CONFIG") != "" {
			configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")
		}

		if deleteNamespace == "" || resourceType == "" || resourceName == "" {
			logrus.Fatalf("Namespace, resource type, and resource name are required")
		}

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

					k8s.DeleteResource(clientset, name, deleteNamespace, resourceType, resourceName)
				}(name, context)
			}
		}
		wg.Wait()
	},
}
