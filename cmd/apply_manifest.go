package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"multicluster/config"
	"multicluster/k8s"
	"multicluster/utils"
	"os"
	"sync"
)

var (
	manifestFile string
)

func init() {
	rootCmd.AddCommand(applyManifestCmd)

	applyManifestCmd.PersistentFlags().StringVarP(&configFile, "config", "c", utils.SetConfigDefault(), "Path to the clusters configuration file")
	applyManifestCmd.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", utils.DefaultKubeconfigPath(), "Path to the kubeconfig file")
	applyManifestCmd.PersistentFlags().StringVarP(&manifestFile, "file", "f", "", "Path to the manifest file")
	applyManifestCmd.PersistentFlags().StringVarP(&clusterName, "cluster", "l", "", "Name of the cluster (if empty, apply to all clusters)")
}

var applyManifestCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a manifest file to one or all clusters",
	Run: func(cmd *cobra.Command, args []string) {

		if !cmd.Flags().Changed("config") || os.Getenv("MC_CONFIG") != "" {
			configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")
		}

		if manifestFile == "" {
			logrus.Fatalf("Manifest file is required")
		}

		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			logrus.Fatalf("Failed to load config: %v", err)
		}

		data, err := ioutil.ReadFile(manifestFile)
		if err != nil {
			logrus.Fatalf("Failed to read manifest file: %v", err)
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
					clientset, dynamicClient, err := k8s.CreateK8sClientFromContext(contex, kubeconfigPath)
					if err != nil {
						logrus.Errorf("Failed to create client for cluster %s: %v", name, err)
						return
					}

					if err := k8s.ApplyManifest(clientset, dynamicClient, name, data); err != nil {
						logrus.Errorf("Failed to apply manifest in cluster %s: %v", name, err)
					}
				}(name, contex)
			}
		}
		wg.Wait()
	},
}
