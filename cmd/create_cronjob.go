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
	createCronJobNamespace string
	createCronJobFile      string
)

var createCronJobCmd = &cobra.Command{
	Use:   "create cronjob -f FILE",
	Short: "Cria um cronjob em clusters (kubectl create cronjob ... multi-cluster)",
	Args:  cobra.NoArgs,
	Run:   runCreateCronJobUniversal,
}

func init() {
	createCronJobCmd.Flags().StringVarP(&createCronJobNamespace, "namespace", "n", "", "Namespace alvo")
	createCronJobCmd.Flags().StringVarP(&createCronJobFile, "file", "f", "", "Manifest(YAML) de cronjob")
	rootCmd.AddCommand(createCronJobCmd)
}

func runCreateCronJobUniversal(cmd *cobra.Command, _ []string) {
	if createCronJobFile == "" {
		fmt.Fprintln(os.Stderr, "Obrigatório --file com manifesto CronJob")
		os.Exit(1)
	}
	cronJobBytes, err := ioutil.ReadFile(createCronJobFile)
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
	handler := &handlers.CreateCronJobHandler{
		Namespace:  createCronJobNamespace,
		Manifest:   cronJobBytes,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, clusters, kubeconfigPath, 5)
	utils.OutputCreateCronJob(handler.Results)
}
