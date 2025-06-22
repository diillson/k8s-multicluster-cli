package cmd

import (
	"context"
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/handlers"
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
	"sigs.k8s.io/yaml"
)

var (
	editNamespace string
)

var editCmd = &cobra.Command{
	Use:   "edit RESOURCE NAME",
	Short: "Edita qualquer recurso ao estilo kubectl (multi-cluster)",
	Args:  cobra.ExactArgs(2),
	Run:   runEditUniversal,
}

func init() {
	editCmd.Flags().StringVarP(&editNamespace, "namespace", "n", "", "Namespace, se aplicável")
	rootCmd.AddCommand(editCmd)
}

func runEditUniversal(cmd *cobra.Command, args []string) {
	resource, name := args[0], args[1]
	configFile = utils.ValidateConfig(cmd, "config-file", "MC_CONFIG")
	apiGroup, apiVersion, resourceName := handlers.InferResourceFromInput(resource)

	// 1. Busca (GET) o resource do cluster origem (pega só do primeiro cluster por padrão)
	cfg, _ := config.LoadConfig(configFile)
	contextMap, _ := config.GetContextMap(cfg)
	var cluster executor.ClusterInfo
	for cname, context := range contextMap {
		if clusterName == "" || cname == clusterName {
			cluster = executor.ClusterInfo{Name: cname, Context: context}
			break
		}
	}
	handler := &handlers.EditGetHandler{
		Resource:   resourceName,
		ApiGroup:   apiGroup,
		ApiVersion: apiVersion,
		Namespace:  editNamespace,
		Name:       name,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), handler, []executor.ClusterInfo{cluster}, kubeconfigPath, 1)
	obj := handler.ResultObj
	if obj == nil {
		fmt.Fprintln(os.Stderr, "Não foi possível localizar o recurso.")
		os.Exit(1)
	}

	// 2. Salva como YAML temporário
	tempfile, err := ioutil.TempFile("", fmt.Sprintf("mc-edit-%s-%s-*.yaml", resource, name))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro criando temp file: %v", err)
		os.Exit(1)
	}
	defer os.Remove(tempfile.Name())
	yamldata, _ := yaml.Marshal(obj.Object)
	tempfile.Write(yamldata)
	tempfile.Close()

	// 3. Abre o editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmdEditor := exec.Command(editor, tempfile.Name())
	cmdEditor.Stdin = os.Stdin
	cmdEditor.Stdout = os.Stdout
	cmdEditor.Stderr = os.Stderr
	cmdEditor.Run()

	// 4. Lê de volta
	edited, err := ioutil.ReadFile(tempfile.Name())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro lendo o arquivo: ", err)
		os.Exit(1)
	}
	var newObj map[string]interface{}
	if err := yaml.Unmarshal(edited, &newObj); err != nil {
		fmt.Fprintln(os.Stderr, "YAML inválido:", err)
		os.Exit(1)
	}

	// 5. PATCH em todos clusters
	clusters := []executor.ClusterInfo{}
	for cname, context := range contextMap {
		if clusterName == "" || cname == clusterName {
			clusters = append(clusters, executor.ClusterInfo{Name: cname, Context: context})
		}
	}
	updateHandler := &handlers.EditUpdateHandler{
		Resource:   resourceName,
		ApiGroup:   apiGroup,
		ApiVersion: apiVersion,
		Namespace:  editNamespace,
		Name:       name,
		NewObj:     newObj,
		KubeConfig: kubeconfigPath,
	}
	executor.ExecuteOnClusters(context.Background(), updateHandler, clusters, kubeconfigPath, 5)
	utils.OutputEdit(updateHandler.Results)
}
