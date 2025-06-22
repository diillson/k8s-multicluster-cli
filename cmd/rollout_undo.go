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
	rolloutUndoNamespace string
	rolloutUndoRevision  int64
)

var rolloutUndoCmd = &cobra.Command{
	Use:   "rollout undo RESOURCE NAME [--to-revision=N]",
	Short: "Rollback de revisões (rollout undo multi-cluster)",
	Args:  cobra.MinimumNArgs(3),
	Run:   runRolloutUndoUniversal,
}

func init() {
	rolloutUndoCmd.Flags().StringVarP(&rolloutUndoNamespace, "namespace", "n", "", "Namespace do recurso")
	rolloutUndoCmd.Flags().Int64Var(&rolloutUndoRevision, "to-revision", 0, "Número da revisão para fazer rollback (default: anterior)")
	rootCmd.AddCommand(rolloutUndoCmd)
}

func runRolloutUndoUniversal(cmd *cobra.Command, args []string) {
	subcmd := args[0]
	resource := args[1]
	name := args[2]
	if subcmd != "undo" {
		fmt.Println("Uso: rollout undo RESOURCE NAME [--to-revision=N]")
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
	handler := &handlers.RolloutUndoHandler{
		Resource:   resourceName,
		ApiGroup:   apiGroup,
		ApiVersion: apiVersion,
		Namespace:  rolloutUndoNamespace,
		Name:       name,
		Revision:   rolloutUndoRevision,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputRolloutUndo(handler.Results)
}
