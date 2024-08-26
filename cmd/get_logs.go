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
	logNamespaces []string
	podNames      []string
	clusters      []string
	container     string
	follow        bool
)

func init() {

	getLogsCmd.PersistentFlags().StringVarP(&configFile, "config", "c", utils.SetConfigDefault(), "Path to the clusters configuration file")
	getLogsCmd.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", utils.DefaultKubeconfigPath(), "Path to the kubeconfig file")
	getLogsCmd.PersistentFlags().StringSliceVarP(&logNamespaces, "namespace", "n", nil, "Namespaces of the pods")
	getLogsCmd.PersistentFlags().StringSliceVarP(&podNames, "pod", "p", nil, "Names of the pods")
	getLogsCmd.PersistentFlags().StringSliceVarP(&clusters, "cluster", "l", nil, "Names of the clusters")
	getLogsCmd.PersistentFlags().StringVarP(&container, "container", "t", "", "Name of the container (optional)")
	getLogsCmd.PersistentFlags().BoolVarP(&follow, "follow", "f", false, "Follow the logs in real time")
}

var getLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get logs from specific pods",
	Run: func(cmd *cobra.Command, args []string) {
		configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")

		if len(logNamespaces) == 0 || len(podNames) == 0 || len(clusters) == 0 {
			logrus.Fatalf("Clusters, Namespaces, and pod names are required")
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
		for _, cluster := range clusters {
			for _, namespace := range logNamespaces {
				for _, pod := range podNames {
					if context, exists := contextMap[cluster]; exists {
						wg.Add(1)
						go func(cluster, namespace, pod, context string) {
							defer wg.Done()
							clientset, _, err := k8s.CreateK8sClientFromContext(context, kubeconfigPath)
							if err != nil {
								logrus.Errorf("Failed to create client for cluster %s: %v", cluster, err)
								return
							}

							k8s.GetPodLogs(clientset, cluster, namespace, pod, container, follow)
						}(cluster, namespace, pod, context)
					} else {
						logrus.Errorf("Cluster %s not found in context map", cluster)
					}
				}
			}
		}
		wg.Wait()
	},
}
