package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/handlers"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/spf13/cobra"
)

var (
	createJobNamespace string
	createJobFile      string
)

var createJobCmd = &cobra.Command{
	Use:   "create job -f FILE",
	Short: "Cria um job em clusters (kubectl create job ... multi-cluster)",
	Args:  cobra.NoArgs,
	Run:   runCreateJobUniversal,
}

func init() {
	createJobCmd.Flags().StringVarP(&createJobNamespace, "namespace", "n", "", "Namespace alvo")
	createJobCmd.Flags().StringVarP(&createJobFile, "file", "f", "", "Manifest(YAML) de Job")
	rootCmd.AddCommand(createJobCmd)
}

func runCreateJobUniversal(cmd *cobra.Command, _ []string) {
	if createJobFile == "" {
		fmt.Fprintln(os.Stderr, "Obrigatório --file com manifesto Job")
		os.Exit(1)
	}
	jobBytes, err := ioutil.ReadFile(createJobFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro lendo arquivo: %v", err)
		os.Exit(1)
	}
	configFile = utils.ValidateConfig(cmd, "config-file", "MC_CONFIG")

	cfg, _ := config.LoadConfig(configFile)
	contextMap, _ := config.GetContextMap(cfg)
	clusters := []executor.ClusterInfo{}
	for cname, context := range contextMap {
		if clusterName == "" || cname == clusterName {
			clusters = append(clusters, executor.ClusterInfo{Name: cname, Context: context})
		}
	}
	handler := &handlers.CreateJobHandler{
		Namespace:  createJobNamespace,
		Manifest:   jobBytes,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputCreateJob(handler.Results)
}
