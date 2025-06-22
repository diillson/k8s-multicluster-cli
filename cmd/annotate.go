package cmd

import (
	"context"
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/handlers"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/spf13/cobra"
	"strings"
)

var (
	annotateNamespace string
)

var annotateCmd = &cobra.Command{
	Use:   "annotate RESOURCE NAME KEY=VAL [KEY=VAL]...",
	Short: "Adiciona/altera annotations em recursos em múltiplos clusters (kubectl annotate)",
	Args:  cobra.MinimumNArgs(3),
	Run:   runAnnotateUniversal,
}

func init() {
	annotateCmd.Flags().StringVarP(&annotateNamespace, "namespace", "n", "", "Namespace para recursos namespaced")
	rootCmd.AddCommand(annotateCmd)
}

func runAnnotateUniversal(cmd *cobra.Command, args []string) {
	resource := args[0]
	name := args[1]
	annotations := args[2:]
	annMap := make(map[string]string)
	for _, a := range annotations {
		kv := strings.SplitN(a, "=", 2)
		if len(kv) != 2 {
			fmt.Printf("Annotação inválida: %s\n", a)
			return
		}
		annMap[kv[0]] = kv[1]
	}
	configFile = utils.ValidateConfig(cmd, "config-file", "MC_CONFIG")
	apiGroup, apiVersion, resourceName := handlers.InferResourceFromInput(resource)
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
	handler := &handlers.AnnotateHandler{
		Resource:    resourceName,
		ApiGroup:    apiGroup,
		ApiVersion:  apiVersion,
		Namespace:   annotateNamespace,
		Name:        name,
		Annotations: annMap,
		KubeConfig:  kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputAnnotate(handler.Results)
}
