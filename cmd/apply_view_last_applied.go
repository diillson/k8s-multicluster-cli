package cmd

import (
	"context"

	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/handlers"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/spf13/cobra"
)

var (
	viewLastAppliedNamespace string
)

var viewLastAppliedCmd = &cobra.Command{
	Use:   "apply view-last-applied RESOURCE NAME",
	Short: "Exibe annotation 'last-applied-config' (multi-cluster, kubectl-style)",
	Args:  cobra.ExactArgs(3),
	Run:   runViewLastAppliedUniversal,
}

func init() {
	viewLastAppliedCmd.Flags().StringVarP(&viewLastAppliedNamespace, "namespace", "n", "", "Namespace")
	rootCmd.AddCommand(viewLastAppliedCmd)
}

func runViewLastAppliedUniversal(cmd *cobra.Command, args []string) {
	resource := args[0]
	name := args[1]
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
	handler := &handlers.ViewLastAppliedHandler{
		Resource:   resourceName,
		ApiGroup:   apiGroup,
		ApiVersion: apiVersion,
		Namespace:  viewLastAppliedNamespace,
		Name:       name,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputViewLastApplied(handler.Results)
}
