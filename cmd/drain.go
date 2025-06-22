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
	drainForce        bool
	drainIgnoreDaemon bool
	drainDeleteData   bool
	drainTimeout      int
	drainNamespace    string
)

var drainCmd = &cobra.Command{
	Use:   "drain NODE [flags]",
	Short: "Esvazia e drena um node para manutenção/multicluster (kubectl drain)",
	Args:  cobra.ExactArgs(1),
	Run:   runDrainUniversal,
}

func init() {
	drainCmd.Flags().BoolVar(&drainForce, "force", false, "Forçar drain mesmo se houver pods não tolerados")
	drainCmd.Flags().BoolVar(&drainIgnoreDaemon, "ignore-daemonsets", true, "Ignorar pods daemonset")
	drainCmd.Flags().BoolVar(&drainDeleteData, "delete-emptydir-data", false, "Deletar emptyDir local ao drenar")
	drainCmd.Flags().IntVar(&drainTimeout, "timeout", 600, "Timeout para drain (segundos)")
	drainCmd.Flags().StringVarP(&drainNamespace, "namespace", "n", "", "Namespace/alvo (opcional)")
	rootCmd.AddCommand(drainCmd)
}

func runDrainUniversal(cmd *cobra.Command, args []string) {
	node := args[0]
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
	handler := &handlers.DrainHandler{
		Node:         node,
		Force:        drainForce,
		IgnoreDaemon: drainIgnoreDaemon,
		DeleteData:   drainDeleteData,
		Timeout:      drainTimeout,
		Namespace:    drainNamespace,
		KubeConfig:   kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 1)
	utils.OutputDrain(handler.Results)
}
