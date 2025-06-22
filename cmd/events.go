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

var eventsNamespace string

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Lista eventos de todos clusters (kubectl events universal)",
	Run:   runEventsUniversal,
}

func init() {
	eventsCmd.Flags().StringVarP(&eventsNamespace, "namespace", "n", "", "Namespace dos eventos (default: todos)")
	rootCmd.AddCommand(eventsCmd)
}

func runEventsUniversal(cmd *cobra.Command, _ []string) {
	configFile = utils.ValidateConfig(cmd, "config-file", "MC_CONFIG")
	namespaces := utils.GetNamespaces(eventsNamespace)

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

	handler := &handlers.EventsHandler{
		Namespaces: namespaces,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputEvents(handler.Results)
}
