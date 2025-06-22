package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/handlers"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/spf13/cobra"
)

var (
	rolloutNamespace string
	rolloutTimeout   int
)

var rolloutStatusCmd = &cobra.Command{
	Use:   "rollout status RESOURCE NAME",
	Short: "Verifica status do rollout de um recurso (deployment/statefulset/...) multi-cluster",
	Args:  cobra.ExactArgs(3),
	Run:   runRolloutStatusUniversal,
}

var rolloutRestartCmd = &cobra.Command{
	Use:   "rollout restart RESOURCE NAME",
	Short: "Restart deployment/statefulset/etc forçando novo rollout",
	Args:  cobra.ExactArgs(3),
	Run:   runRolloutRestartUniversal,
}

func init() {
	rolloutStatusCmd.Flags().StringVarP(&rolloutNamespace, "namespace", "n", "", "Namespace do recurso")
	rolloutStatusCmd.Flags().IntVar(&rolloutTimeout, "timeout", 300, "Timeout em segundos para esperar rollout")
	rootCmd.AddCommand(rolloutStatusCmd)
}

func runRolloutStatusUniversal(cmd *cobra.Command, args []string) {
	subcmd := args[0]
	resource := args[1]
	name := args[2]

	if strings.ToLower(subcmd) != "status" {
		fmt.Printf("Uso: rollout status RESOURCE NAME\n")
		return
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
	handler := &handlers.RolloutStatusHandler{
		Resource:   resourceName,
		ApiGroup:   apiGroup,
		ApiVersion: apiVersion,
		Namespace:  rolloutNamespace,
		Name:       name,
		Timeout:    rolloutTimeout,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputRolloutStatus(handler.Results)
}

func runRolloutRestartUniversal(cmd *cobra.Command, args []string) {
	subcmd := args[0]
	resource := args[1]
	name := args[2]
	if strings.ToLower(subcmd) != "restart" {
		fmt.Println("Uso: rollout restart RESOURCE NAME")
		return
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
	handler := &handlers.RolloutRestartHandler{
		Resource:   resourceName,
		ApiGroup:   apiGroup,
		ApiVersion: apiVersion,
		Namespace:  rolloutNamespace,
		Name:       name,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputRolloutRestart(handler.Results)
}
