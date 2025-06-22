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
	pfNamespace string
	pfAddress   string
	pfPorts     []string
	pfPod       string
	pfService   string
)

var portForwardCmd = &cobra.Command{
	Use:   "port-forward (--pod POD|--service SERVICE) [LOCAL_PORT:]REMOTE_PORT...",
	Short: "Port-forward para pods/services em múltiplos clusters ao mesmo tempo",
	Run:   runPortForwardUniversal,
}

func init() {
	portForwardCmd.Flags().StringVarP(&pfNamespace, "namespace", "n", "", "Namespace dos recursos")
	portForwardCmd.Flags().StringVar(&pfAddress, "address", "127.0.0.1", "Local address to bind (default 127.0.0.1)")
	portForwardCmd.Flags().StringVar(&pfPod, "pod", "", "Nome do pod alvo")
	portForwardCmd.Flags().StringVar(&pfService, "service", "", "Nome do service alvo (exclusivo com --pod)")
	portForwardCmd.Flags().StringSliceVar(&pfPorts, "ports", nil, "Portas no formato [LOCAL_PORT:]REMOTE_PORT")
	rootCmd.AddCommand(portForwardCmd)
}

func runPortForwardUniversal(cmd *cobra.Command, _ []string) {
	configFile = utils.ValidateConfig(cmd, "config-file", "MC_CONFIG")
	if pfPod == "" && pfService == "" {
		fmt.Println("Obrigatório informar --pod POD ou --service SERVICE")
		os.Exit(1)
	}
	if pfPod != "" && pfService != "" {
		fmt.Println("Use apenas um dos dois: --pod ou --service")
		os.Exit(1)
	}
	if len(pfPorts) == 0 {
		fmt.Println("Informe pelo menos uma porta em --ports ex: --ports 8000:80")
		os.Exit(1)
	}
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

	handler := &handlers.PortForwardHandler{
		Resource:   "pod",
		Name:       pfPod,
		Namespace:  pfNamespace,
		Address:    pfAddress,
		Ports:      pfPorts,
		KubeConfig: kubeconfigPath,
	}
	if pfService != "" {
		handler.Resource = "service"
		handler.Name = pfService
	}

	executor.ExecuteOnClusters(
		context.Background(),
		handler,
		clusters,
		kubeconfigPath,
		1, // Por padrão, port-forward é melhor ser sequencial; múltiplos podem ser caóticos
	)
	utils.OutputPortForward(handler.Results)
}
