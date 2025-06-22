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

var (
	rolloutHistoryNamespace string
)

var rolloutHistoryCmd = &cobra.Command{
	Use:   "rollout history RESOURCE NAME",
	Short: "Mostra revisões (history) de deployments/setfulsets multi-cluster",
	Args:  cobra.ExactArgs(3),
	Run:   runRolloutHistoryUniversal,
}

func init() {
	rolloutHistoryCmd.Flags().StringVarP(&rolloutHistoryNamespace, "namespace", "n", "", "Namespace do recurso")
	rootCmd.AddCommand(rolloutHistoryCmd)
}

func runRolloutHistoryUniversal(cmd *cobra.Command, args []string) {
	subcmd := args[0]
	resource := args[1]
	name := args[2]
	if subcmd != "history" {
		fmt.Println("Uso: rollout history RESOURCE NAME")
		return
	}
	configFile = utils.ValidateConfig(cmd, "config-file", "MC_CONFIG")
	apiGroup, apiVersion, resourceName := handlers.InferResourceFromInput(resource)
	cfg, _ := config.LoadConfig(configFile)
	contextMap, _ := config.GetContextMap(cfg)
	clusters := []executor.ClusterInfo{}
	for cname, context := range contextMap {
		if clusterName == "" || cname == clusterName {
			clusters = append(clusters, executor.ClusterInfo{Name: cname, Context: context})
		}
	}
	handler := &handlers.RolloutHistoryHandler{
		Resource:   resourceName,
		ApiGroup:   apiGroup,
		ApiVersion: apiVersion,
		Namespace:  rolloutHistoryNamespace,
		Name:       name,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputRolloutHistory(handler.Results)
}
