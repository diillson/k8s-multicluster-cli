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
	logsNamespace string
	logsContainer string
	follow        bool
	tail          int64
	sinceSeconds  int64
)

var logsCmd = &cobra.Command{
	Use:   "logs POD",
	Short: "Exibe logs de Pods (universal multicluster)",
	Args:  cobra.MinimumNArgs(1),
	Run:   runLogsUniversal,
}

func init() {
	logsCmd.Flags().StringVarP(&logsNamespace, "namespace", "n", "", "Namespace do Pod (default: default)")
	logsCmd.Flags().StringVarP(&logsContainer, "container", "c", "", "Container alvo (para pods multi-container)")
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Stream follow (como tail -f)")
	logsCmd.Flags().Int64Var(&tail, "tail", 100, "Número de linhas finais (default 100)")
	logsCmd.Flags().Int64Var(&sinceSeconds, "since", 0, "Mostrar logs desde N segundos atrás")
	rootCmd.AddCommand(logsCmd)
}

func runLogsUniversal(cmd *cobra.Command, args []string) {
	podName := args[0]
	configFile = utils.ValidateConfig(cmd, "config", "MC_CONFIG")
	namespaces := utils.GetNamespaces(logsNamespace)

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
	handler := &handlers.LogsHandler{
		PodName:    podName,
		Namespaces: namespaces,
		Container:  logsContainer,
		Follow:     follow,
		Tail:       tail,
		Since:      sinceSeconds,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(
		context.Background(),
		handler,
		clusters,
		kubeconfigPath,
		5,
	)
	utils.OutputLogs(handler.Results)
}
