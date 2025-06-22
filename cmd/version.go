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
	"runtime"
)

var (
	shortVersion  bool
	clientVersion = "2.0.0" // Ou gere via -ldflags na build
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Mostra as versões do CLI e (opcionalmente) do servidor Kubernetes",
	Run:   runVersionUniversal,
}

func init() {
	versionCmd.Flags().BoolVar(&shortVersion, "short", false, "Mostrar só versão sem detalhes")
	rootCmd.AddCommand(versionCmd)
}

func runVersionUniversal(cmd *cobra.Command, _ []string) {
	// CLIENT
	fmt.Printf("CLI Version: %s\n", clientVersion)
	fmt.Printf("GoVersion:   %s\n", runtime.Version())
	fmt.Printf("Platform:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Compiler:    %s\n", runtime.Compiler)
	if shortVersion {
		os.Exit(0)
	}

	// SERVIDOR
	configFile = utils.ValidateConfig(cmd, "config-file", "MC_CONFIG")
	cfg, _ := config.LoadConfig(configFile)
	contextMap, _ := config.GetContextMap(cfg)
	clusters := []executor.ClusterInfo{}
	for cname, context := range contextMap {
		if clusterName == "" || cname == clusterName {
			clusters = append(clusters, executor.ClusterInfo{Name: cname, Context: context})
		}
	}
	handler := &handlers.VersionHandler{
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputVersion(handler.Results)
}
