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
	caniNamespace string
)

var caniCmd = &cobra.Command{
	Use:   "can-i VERBO RECURSO [NAME]",
	Short: "Verifica permissão para verbo/recurso em todos os clusters (kubectl auth can-i)",
	Args:  cobra.MinimumNArgs(2),
	Run:   runCanIUniversal,
}

func init() {
	caniCmd.Flags().StringVarP(&caniNamespace, "namespace", "n", "", "Namespace onde testar permissão (opcional)")
	rootCmd.AddCommand(caniCmd)
}

func runCanIUniversal(cmd *cobra.Command, args []string) {
	configFile = utils.ValidateConfig(cmd, "config-file", "MC_CONFIG")
	verb, resource := args[0], args[1]
	var name string
	if len(args) > 2 {
		name = args[2]
	}
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
	handler := &handlers.CanIHandler{
		Verb:       verb,
		Resource:   resource,
		Name:       name,
		Namespace:  caniNamespace,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputCanI(handler.Results)
}
