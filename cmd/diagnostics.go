package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/handlers"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/spf13/cobra"
)

var diagFast bool

var diagnosticsCmd = &cobra.Command{
	Use:   "diagnostics",
	Short: "Self-test da CLI e dos clusters (sanity-check, health, conexão, auth, version, api-resources)",
	Run:   runDiagnostics,
}

func init() {
	diagnosticsCmd.Flags().BoolVar(&diagFast, "fast", false, "Roda apenas checks rápidos (conexão/version)")
	rootCmd.AddCommand(diagnosticsCmd)
}

func runDiagnostics(cmd *cobra.Command, _ []string) {
	fmt.Println("🔍 Iniciando testes e diagnóstico dos clusters...")
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

	okCh := make(chan bool)
	time.AfterFunc(5*time.Second, func() {
		okCh <- true
	})

	// VERIFICA VERSÃO
	fmt.Println("\n✔️  Testando acesso à API (descobrindo versão/saúde):")
	vhandler := &handlers.VersionHandler{KubeConfig: kubeconfigPath}
	executor.ExecuteOnClusters(context.Background(), vhandler, clusters, kubeconfigPath, 5)
	utils.OutputVersion(vhandler.Results)

	if diagFast {
		return
	}

	// VERIFICA API-RESOURCES (CRDs e discovery)
	fmt.Println("\n✔️  Testando discovery de recursos (api-resources):")
	arHandler := &handlers.APIResourcesHandler{KubeConfig: kubeconfigPath}
	executor.ExecuteOnClusters(context.Background(), arHandler, clusters, kubeconfigPath, 5)
	utils.OutputAPIResources(arHandler.Results)

	// CHECAGEM DE NODES
	fmt.Println("\n✔️  Listando nodes (verifica RBAC, auth, resposta):")
	nodesHandler := &handlers.GetResourceHandler{
		Resource:   "nodes",
		ApiGroup:   "",
		ApiVersion: "v1",
		Namespaces: []string{""},
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), nodesHandler, clusters, kubeconfigPath, 5)
	utils.OutputGeneric(nodesHandler.Results, "table", "nodes")

	fmt.Println("\n🏁 Diagnóstico concluído.")
}
