package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/handlers"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/spf13/cobra"
)

var (
	setLastAppliedNamespace string
	setLastAppliedFile      string
)

var setLastAppliedCmd = &cobra.Command{
	Use:   "apply set-last-applied RESOURCE NAME -f FILE",
	Short: "Salva manifest como annotation 'last-applied-config' (multi-cluster, kubectl-style)",
	Args:  cobra.ExactArgs(3),
	Run:   runSetLastAppliedUniversal,
}

func init() {
	setLastAppliedCmd.Flags().StringVarP(&setLastAppliedNamespace, "namespace", "n", "", "Namespace")
	setLastAppliedCmd.Flags().StringVarP(&setLastAppliedFile, "file", "f", "", "Manifest file (obrigatório)")
	rootCmd.AddCommand(setLastAppliedCmd)
}

func runSetLastAppliedUniversal(cmd *cobra.Command, args []string) {
	resource := args[0]
	name := args[1]
	if setLastAppliedFile == "" {
		fmt.Fprintln(os.Stderr, "Obrigatório informar --file FILE com o manifesto")
		os.Exit(1)
	}
	manifestBytes, err := ioutil.ReadFile(setLastAppliedFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler arquivo: %v", err)
		os.Exit(1)
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
	handler := &handlers.SetLastAppliedHandler{
		Resource:   resourceName,
		ApiGroup:   apiGroup,
		ApiVersion: apiVersion,
		Namespace:  setLastAppliedNamespace,
		Name:       name,
		Manifest:   manifestBytes,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputSetLastApplied(handler.Results)
}
