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
	describeNamespace string
)

var describeCmd = &cobra.Command{
	Use:   "describe [RESOURCE] [NAME]",
	Short: "Describe qualquer recurso em todos clusters",
	Args:  cobra.MinimumNArgs(2),
	Run:   runDescribeUniversal,
}

func init() {

	describeCmd.Flags().StringVarP(&describeNamespace, "namespace", "n", "", "Namespace (default: all/all-namespaces)")
	rootCmd.AddCommand(describeCmd)
}

func runDescribeUniversal(cmd *cobra.Command, args []string) {
	resource := args[0]
	name := args[1]
	configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")
	apiGroup, apiVersion, resourceName := handlers.InferResourceFromInput(resource)
	namespaces := utils.GetNamespaces(describeNamespace)
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}
	contextMap, err := config.GetContextMap(cfg)
	if err != nil {
		fmt.Printf("Failed to get context map: %v\n", err)
		return
	}
	clusters := []executor.ClusterInfo{}
	for cname, context := range contextMap {
		if clusterName == "" || cname == clusterName {
			clusters = append(clusters, executor.ClusterInfo{Name: cname, Context: context})
		}
	}
	handler := &handlers.DescribeResourceHandler{
		Resource:   resourceName,
		ApiGroup:   apiGroup,
		ApiVersion: apiVersion,
		Namespaces: namespaces,
		Name:       name,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputDescribe(handler.Results)
}
