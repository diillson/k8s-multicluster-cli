package cmd

import (
	"fmt"
	"os"

	"github.com/diillson/k8s-multicluster-cli/config"
	"github.com/diillson/k8s-multicluster-cli/utils"
	"github.com/spf13/cobra"
)

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Consulta e altera configurações dos clusters (kubectl-style)",
	}

	viewCmd = &cobra.Command{
		Use:   "view",
		Short: "Mostra o conteúdo do arquivo de configuração multicluster (clusters.json/config.json)",
		Run: func(cmd *cobra.Command, args []string) {
			v, err := utils.ReadFileOrExit(configFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao ler config: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(v)
		},
	}

	getContextsCmd = &cobra.Command{
		Use:   "get-contexts",
		Short: "Lista cluster/contextos definidos no seu config.json",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadConfig(configFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao ler config: %v\n", err)
				os.Exit(1)
			}
			for _, cl := range cfg.Clusters {
				fmt.Printf("%-20s %s\n", cl.Name, cl.Context)
			}
		},
	}

	useContextCmd = &cobra.Command{
		Use:   "use-context NAME",
		Short: "Seta ou troca o contexto default (não altera o kubeconfig, apenas o multicluster)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			cfg, err := config.LoadConfig(configFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao ler config: %v\n", err)
				os.Exit(1)
			}
			found := false
			for _, cl := range cfg.Clusters {
				if cl.Name == name {
					fmt.Printf("Contexto '%s' setado.\n", name)
					utils.SetDefaultContext(name)
					found = true
				}
			}
			if !found {
				fmt.Fprintf(os.Stderr, "Contexto '%s' não encontrado!\n", name)
				os.Exit(1)
			}
		},
	}
)

func init() {
	configCmd.AddCommand(viewCmd)
	configCmd.AddCommand(getContextsCmd)
	configCmd.AddCommand(useContextCmd)
	rootCmd.AddCommand(configCmd)
}
