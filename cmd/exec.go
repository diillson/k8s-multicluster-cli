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
	execNamespace string
	execContainer string
	tty           bool
	stdin         bool
)

var execCmd = &cobra.Command{
	Use:   "exec POD -- COMMAND [args...]",
	Short: "Executa comando em container (universal multicluster)",
	Args:  cobra.MinimumNArgs(2),
	Run:   runExecUniversal,
}

func init() {
	execCmd.Flags().StringVarP(&execNamespace, "namespace", "n", "", "Namespace do Pod")
	execCmd.Flags().StringVarP(&execContainer, "container", "c", "", "Container alvo")
	execCmd.Flags().BoolVarP(&tty, "tty", "t", false, "Forçar TTY (interativo)")
	execCmd.Flags().BoolVarP(&stdin, "stdin", "i", false, "Usar stdin")
	rootCmd.AddCommand(execCmd)
}

func runExecUniversal(cmd *cobra.Command, args []string) {
	podName := args[0]
	command := args[1:]
	configFile = utils.ValidateConfig(cmd, "config-file", "MC_CONFIG")
	namespaces := utils.GetNamespaces(execNamespace)

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

	// ✍️ Recusa exec -i/--stdin ou -t/--tty em múltiplos clusters
	if (stdin || tty) && len(clusters) > 1 {
		fmt.Println("Execução interativa (-i/-t) suportada apenas para 1 cluster alvo! Use --cluster para especificar.")
		os.Exit(1)
	}

	handler := &handlers.ExecHandler{
		PodName:    podName,
		Namespaces: namespaces,
		Container:  execContainer,
		Command:    command,
		Stdin:      stdin,
		TTY:        tty,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(
		context.Background(),
		handler,
		clusters,
		kubeconfigPath,
		1, // interativo: só um por vez
	)
	utils.OutputExec(handler.Results)
}
