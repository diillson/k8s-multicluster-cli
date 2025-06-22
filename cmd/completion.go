package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Gera scripts de auto-complete para shell (kubectl-style)",
	Long: `
    Para ativar o auto-complete, siga as instruções conforme seu shell (igual kubectl):
    
      Bash:
        source <(multicluster completion bash)
        # ou
        multicluster completion bash > ~/.bash_completion.d/multicluster
    
      ZSH:
        source <(multicluster completion zsh)
        # ou
        multicluster completion zsh > "${fpath[1]}/_multicluster"
    
      Fish:
        multicluster completion fish | source
        # ou
        multicluster completion fish > ~/.config/fish/completions/multicluster.fish
    
      PowerShell:
        multicluster completion powershell | Out-String | Invoke-Expression
    `,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			// ZSH completion cobre bug do cobra (ver kubectl source)
			fmt.Fprintf(os.Stdout, "#compdef multicluster\ncompdef _multicluster multicluster\n")
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			fmt.Fprintf(os.Stderr, "Shell não suportado %s\n", args[0])
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
