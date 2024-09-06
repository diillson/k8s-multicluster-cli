package cmd

import (
	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/k8s"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	describeCluster   string
	describeNamespace string
	//describePodName   string
)

func init() {

	describePodCmd.PersistentFlags().StringVarP(&configFile, "config", "c", utils.SetConfigDefault(), "Path to the clusters configuration file")
	describePodCmd.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", utils.DefaultKubeconfigPath(), "Path to the kubeconfig file")
	describePodCmd.PersistentFlags().StringVarP(&describeCluster, "cluster", "l", "", "Name of the cluster")
	describePodCmd.PersistentFlags().StringVarP(&describeNamespace, "namespace", "n", "", "Namespace of the pod")
	//describePodCmd.PersistentFlags().StringVarP(&describePodName, "pod", "p", "", "Name of the pod")
}

var describePodCmd = &cobra.Command{
	Use:   "pod [pod_name]",
	Short: "Describe a specific pod in a specific cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !cmd.Flags().Changed("config") || os.Getenv("MC_CONFIG") != "" {
			configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")
		}

		describePodName := args[0]

		if describeCluster == "" || describeNamespace == "" || describePodName == "" {
			logrus.Fatalf("Cluster name, namespace, and pod name are required")
		}

		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			logrus.Fatalf("Failed to load config: %v", err)
		}

		contextMap, err := config.GetContextMap(cfg)
		if err != nil {
			logrus.Fatalf("Failed to get context map from kubeconfig: %v", err)
		}

		var clusterFound bool
		for name, contex := range contextMap {
			if name == describeCluster {
				clusterFound = true
				clientset, _, err := k8s.CreateK8sClientFromContext(contex, kubeconfigPath)
				if err != nil {
					logrus.Fatalf("Failed to create client for cluster %s: %v", name, err)
				}

				if err := k8s.DescribePod(clientset, name, describeNamespace, describePodName); err != nil {
					logrus.Fatalf("Failed to describe pod %s in cluster %s: %v", describePodName, name, err)
				}
				break
			}
		}

		if !clusterFound {
			logrus.Fatalf("Cluster %s not found in the configuration", describeCluster)
		}
	},
}
