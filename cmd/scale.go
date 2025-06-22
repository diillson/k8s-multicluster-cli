package cmd

import (
	"context"
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/handlers"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/spf13/cobra"
	"strconv"
)

var (
	scaleNamespace string
)

var scaleCmd = &cobra.Command{
	Use:   "scale RESOURCE NAME --replicas=N",
	Short: "Altera réplicas de recursos escaláveis em múltiplos clusters (kubectl scale)",
	Args:  cobra.ExactArgs(2),
	Run:   runScaleUniversal,
}

func init() {
	scaleCmd.Flags().StringVarP(&scaleNamespace, "namespace", "n", "", "Namespace do recurso")
	rootCmd.AddCommand(scaleCmd)
}

func runScaleUniversal(cmd *cobra.Command, args []string) {
	resource := args[0]
	name := args[1]
	replicasStr, err := cmd.Flags().GetString("replicas")
	if err != nil || replicasStr == "" {
		fmt.Println("Obrigatório definir --replicas=N")
		return
	}
	replicas, err := strconv.Atoi(replicasStr)
	if err != nil || replicas < 0 {
		fmt.Println("Número de réplicas inválido:", replicasStr)
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
	handler := &handlers.ScaleHandler{
		Resource:   resourceName,
		ApiGroup:   apiGroup,
		ApiVersion: apiVersion,
		Namespace:  scaleNamespace,
		Name:       name,
		Replicas:   int32(replicas),
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputScale(handler.Results)
}

func init() {
	scaleCmd.Flags().String("replicas", "", "Novo número de réplicas (obrigatório)")
}
