package cmd

import "github.com/spf13/cobra"

// Estas variáveis serão usadas pelos subcomandos.
var (
	clusterName    string
	configFile     string
	kubeconfigPath string
)

var rootCmd = &cobra.Command{
	Use:   "multicluster",
	Short: "CLI multi-cluster ao estilo kubectl",
	Long: `Multicluster é um CLI universal, inspirado no kubectl,
    projetado para rodar comandos simultaneamente em múltiplos clusters Kubernetes.
    
    Veja exemplos de uso:
      multicluster get pods --all-namespaces
      multicluster describe node node1 -ct prod
      multicluster port-forward --pod mypod -n myns --ports 8080:80
    
    Comandos especiais:
      plugin         - executa plugins externos kubectl-style
      completion     - gera scripts de autocompletar para bash/zsh/fish/pwsh
      config         - gerencia cluster configs multicluster
    
    Veja a documentação completa em: https://github.com/diillson/k8s-multicluster-cli
    `,
	Example: `
      # Lista pods em todos clusters e namespaces
      multicluster get pods -ct cluster=dev
    
      # Lista services apenas em clusters dev
      multicluster get services -ct cluster=dev`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// FLAGS GLOBAIS
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "", "config.json", "Path do config.json dos clusters")
	rootCmd.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", "", "Path do kubeconfig")
	rootCmd.PersistentFlags().StringVarP(&clusterName, "cluster", "", "", "Cluster alvo (opcional)")

	// SUBCOMANDOS UNIVERSAIS
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(patchCmd)
	rootCmd.AddCommand(describeCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(apiResourcesCmd)
	rootCmd.AddCommand(apiVersionsCmd)
	rootCmd.AddCommand(versionCmd)
}
