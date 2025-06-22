package cmd

import (
	"context"
	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/handlers"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/spf13/cobra"
)

var (
	logsJobNamespace string
	logsJobContainer string
)

var logsJobCmd = &cobra.Command{
	Use:   "logs job JOBNAME",
	Short: "Mostra logs de todos pods controlados por um Job multi-cluster (kubectl logs job/...)",
	Args:  cobra.ExactArgs(2),
	Run:   runLogsJobUniversal,
}

func init() {
	logsJobCmd.Flags().StringVarP(&logsJobNamespace, "namespace", "n", "", "Namespace")
	logsJobCmd.Flags().StringVarP(&logsJobContainer, "container", "c", "", "Container (opcional)")
	rootCmd.AddCommand(logsJobCmd)
}

func runLogsJobUniversal(cmd *cobra.Command, args []string) {
	name := args[1]
	configFile = utils.ValidateConfig(cmd, "config-file", "MC_CONFIG")
	cfg, _ := config.LoadConfig(configFile)
	contextMap, _ := config.GetContextMap(cfg)
	clusters := []executor.ClusterInfo{}
	for cname, context := range contextMap {
		if clusterName == "" || cname == clusterName {
			clusters = append(clusters, executor.ClusterInfo{Name: cname, Context: context})
		}
	}
	handler := &handlers.LogsJobHandler{
		Namespace:  logsJobNamespace,
		JobName:    name,
		Container:  logsJobContainer,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputLogsJob(handler.Results)
}
