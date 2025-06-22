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
	applyFile      string // --file or -f
	applyDirectory string // --directory (opcional, suporte múltiplos arquivos)
	applyNamespace string // --namespace
)

var applyCmd = &cobra.Command{
	Use:   "apply -f FILE [flags]",
	Short: "Apply a manifest file (YAML/JSON) to one or all clusters",
	Args:  cobra.NoArgs, // não precisa de argumento posicional
	Run:   runApplyUniversal,
}

func init() {
	applyCmd.Flags().StringVarP(&applyFile, "file", "f", "", "Manifest file to apply")
	applyCmd.Flags().StringVarP(&applyDirectory, "directory", "d", "", "Directory of manifests to apply")
	applyCmd.Flags().StringVarP(&applyNamespace, "namespace", "n", "", "Namespace target for resources (opcional, sobrescreve no create)")

	rootCmd.AddCommand(applyCmd)
}

func runApplyUniversal(cmd *cobra.Command, _ []string) {
	configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")
	if applyFile == "" && applyDirectory == "" {
		fmt.Println("Obrigatório um --file ou --directory como argumento")
		os.Exit(1)
	}

	var manifests []string
	if applyFile != "" {
		manifests = append(manifests, applyFile)
	}
	if applyDirectory != "" {
		files, err := utils.GatherManifestFiles(applyDirectory)
		if err != nil {
			fmt.Printf("Erro ao ler arquivos do diretório %s: %v\n", applyDirectory, err)
			os.Exit(1)
		}
		manifests = append(manifests, files...)
	}

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
	handler := &handlers.ApplyResourceHandler{
		ManifestFiles: manifests,
		Namespace:     applyNamespace,
		KubeConfig:    kubeconfigPath,
	}
	executor.ExecuteOnClusters(
		context.Background(),
		handler,
		clusters,
		kubeconfigPath,
		5,
	)
	utils.OutputApply(handler.Results)
}
