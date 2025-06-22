package cmd

import (
	"context"
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/handlers"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/spf13/cobra"
	"os"
)

var (
	patchNamespace string
	patchType      string // merge, strategic, json
	patchData      string // pode ser arquivo (@file) ou string
)

var patchCmd = &cobra.Command{
	Use:   "patch [RESOURCE] [NAME] -p PATCH_STRING",
	Short: "Patch any Kubernetes resource in one or all clusters",
	Args:  cobra.MinimumNArgs(2),
	Run:   runPatchUniversal,
}

func init() {
	patchCmd.Flags().StringVarP(&patchNamespace, "namespace", "n", "", "Namespace target")
	patchCmd.Flags().StringVar(&patchType, "type", "strategic", "Patch type (strategic|merge|json)")
	patchCmd.Flags().StringVarP(&patchData, "patch", "p", "", "Patch string payload (ou @path para arquivo)")
	rootCmd.AddCommand(patchCmd)
}

func runPatchUniversal(cmd *cobra.Command, args []string) {
	if patchData == "" {
		fmt.Println("Obrigatório especificar --patch (-p)")
		return
	}
	data := patchData
	if len(patchData) > 0 && patchData[0] == '@' {
		by, err := os.ReadFile(patchData[1:])
		if err != nil {
			fmt.Printf("Erro ao ler arquivo patch: %v\n", err)
			os.Exit(1)
		}
		data = string(by)
	}
	configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")
	resource := args[0]
	name := args[1]
	apiGroup, apiVersion, resourceName := handlers.InferResourceFromInput(resource)
	namespaces := utils.GetNamespaces(patchNamespace)
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
	handler := &handlers.PatchResourceHandler{
		Resource:   resourceName,
		ApiGroup:   apiGroup,
		ApiVersion: apiVersion,
		Namespaces: namespaces,
		Name:       name,
		PatchType:  patchType,
		Patch:      []byte(data),
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputPatch(handler.Results)
}
