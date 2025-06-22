package cmd

import (
	"context"
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/handlers"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/spf13/cobra"
)

var uncordonCmd = &cobra.Command{
	Use:   "uncordon NODE",
	Short: "Marca um node como schedulable (uncordon, kubectl uncordon)",
	Args:  cobra.ExactArgs(1),
	Run:   runUncordonUniversal,
}

func init() {
	rootCmd.AddCommand(uncordonCmd)
}

func runUncordonUniversal(cmd *cobra.Command, args []string) {
	node := args[0]
	configFile = utils.ValidateConfig(cmd, "config-file", "MC_CONFIG")
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Fail loading config: %v\n", err)
		return
	}
	contextMap, err := config.GetContextMap(cfg)
	if err != nil {
		fmt.Printf("Fail loading clusters: %v\n", err)
		return
	}

	clusters := []executor.ClusterInfo{}
	for cname, context := range contextMap {
		if clusterName == "" || cname == clusterName {
			clusters = append(clusters, executor.ClusterInfo{Name: cname, Context: context})
		}
	}
	handler := &handlers.CordonHandler{
		Node:       node,
		Unschedule: false,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 1)
	utils.OutputCordon(handler.Results, "uncordon")
}
