package main

import (
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/cmd"
	"os"
	"os/exec"
)

func main() {
	if err := cmd.Execute(); err != nil {
		// Se o erro for 'unknown command', tente plugin fallback:
		args := os.Args[1:]
		if len(args) > 0 {
			pluginName := "multicluster-" + args[0]
			pluginPath, found := cmd.LookupInPath(pluginName)
			if found {
				proc := exec.Command(pluginPath, args[1:]...)
				proc.Stdin = os.Stdin
				proc.Stdout = os.Stdout
				proc.Stderr = os.Stderr
				if err := proc.Run(); err != nil {
					os.Exit(1)
				}
				os.Exit(0)
			}
		}
		// Se não encontrou plugin, print erro original
		fmt.Fprintf(os.Stderr, "Comando desconhecido ou erro: %v\n", err)
		os.Exit(1)
	}
}
