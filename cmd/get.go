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
	resourceNamespace     string
	resourceOutput        string
	resourceSelector      string
	resourceFieldSelector string
)

var getCmd = &cobra.Command{
	Use:   "get [RESOURCE] [NAME...]",
	Short: "Get any Kubernetes resource from one or all clusters",
	Long:  "Works just like `kubectl get`, but operando em múltiplos clusters simultaneamente.",
	Args:  cobra.MinimumNArgs(1),
	Run:   runGetUniversal,
}

func init() {

	getCmd.Flags().StringVarP(&resourceNamespace, "namespace", "n", "", "Namespace do recurso (padrão: default ou todos os namespaces para recursos namespaced).")
	getCmd.Flags().StringVarP(&resourceOutput, "output", "o", "table", "Formato de saída: 'table' (padrão), 'yaml', 'json'.")
	getCmd.Flags().StringVar(&resourceSelector, "selector", "", "Selector de label para filtrar objetos.")
	getCmd.Flags().StringVar(&resourceFieldSelector, "field-selector", "", "Selector de campo para filtrar objetos.")

	rootCmd.AddCommand(getCmd)
}

func runGetUniversal(cmd *cobra.Command, args []string) {
	configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")
	resource := args[0]
	var names []string
	if len(args) > 1 {
		names = args[1:]
	}

	// Descobre apiGroup/apiVersion padrão via kube discovery, ou mapeias você mesmo. Aqui é simplificado:
	// Exemplo: para pods, group="", version="v1", resource="pods"
	// Para CRDs: group="...", version="...", resource=...

	apiGroup, apiVersion, resourceName := handlers.InferResourceFromInput(resource)
	namespaces := utils.GetNamespaces(resourceNamespace)
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
	for name, context := range contextMap {
		if clusterName == "" || name == clusterName {
			clusters = append(clusters, executor.ClusterInfo{Name: name, Context: context})
		}
	}
	handler := &handlers.GetResourceHandler{
		Resource:      resourceName,
		ApiGroup:      apiGroup,
		ApiVersion:    apiVersion,
		Namespaces:    namespaces,
		KubeConfig:    kubeconfigPath,
		LabelSelector: resourceSelector,
		FieldSelector: resourceFieldSelector,
		Names:         names,
		OutputFormat:  resourceOutput,
	}
	executor.ExecuteOnClusters(
		context.Background(),
		handler,
		clusters,
		kubeconfigPath,
		5,
	)
	utils.OutputGeneric(handler.Results, resourceOutput, resource)
}
