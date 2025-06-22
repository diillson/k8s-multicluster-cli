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
	deleteNamespace string
)

var deleteCmd = &cobra.Command{
	Use:   "delete [RESOURCE] [NAME]",
	Short: "Delete any Kubernetes resource from one or all clusters",
	Args:  cobra.MinimumNArgs(2),
	Run:   runDeleteUniversal,
}

func init() {

	deleteCmd.Flags().StringVarP(&deleteNamespace, "namespace", "n", "", "Namespace do recurso (padrão: default ou todos os namespaces para recursos namespaced).")
	rootCmd.AddCommand(deleteCmd)
}

func runDeleteUniversal(cmd *cobra.Command, args []string) {
	configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")
	resource := args[0]
	name := args[1]

	apiGroup, apiVersion, resourceName := handlers.InferResourceFromInput(resource)
	namespaces := utils.GetNamespaces(deleteNamespace)

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
	handler := &handlers.DeleteResourceHandler{
		Resource:   resourceName,
		ApiGroup:   apiGroup,
		ApiVersion: apiVersion,
		Namespaces: namespaces,
		Name:       name,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(
		context.Background(),
		handler,
		clusters,
		kubeconfigPath,
		5,
	)
	utils.OutputDelete(handler.Results)
}
