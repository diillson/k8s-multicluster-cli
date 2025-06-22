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

var apiVersionsCmd = &cobra.Command{
	Use:   "api-versions",
	Short: "Mostra todas as GROUP/VERSIONS disponíveis nos clusters",
	Run:   runAPIVersions,
}

func init() {
	rootCmd.AddCommand(apiVersionsCmd)
}

func runAPIVersions(cmd *cobra.Command, _ []string) {
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

	handler := &handlers.APIVersionsHandler{
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputAPIVersions(handler.Results)
}
