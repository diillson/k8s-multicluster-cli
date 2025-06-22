package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin [subcomando]",
	Short: "Gerencie e descubra plugins externos multiplataforma (kubectl-style)",
	Long:  "Integração, descoberta, e execução de plugins externos para o multicluster CLI (à la kubectl plugin)",
	Run:   runPluginDiscovery,
}

func init() {
	rootCmd.AddCommand(pluginCmd)
}

func runPluginDiscovery(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		// Lista todos plug-ins disponíveis
		fmt.Println("Plugins disponíveis:\n")
		plugins := discoverPlugins("multicluster-")
		for _, p := range plugins {
			fmt.Println("- " + p)
		}
		return
	}
	pluginName := "multicluster-" + args[0]
	pluginPath, found := LookupInPath(pluginName)
	if !found {
		fmt.Fprintf(os.Stderr, "Plugin não encontrado: %s\n", pluginName)
		os.Exit(1)
	}
	// Executa
	proc := exec.Command(pluginPath, args[1:]...)
	proc.Stdin = os.Stdin
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	if err := proc.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Falha ao rodar plugin '%s': %v\n", pluginName, err)
		os.Exit(1)
	}
}

func discoverPlugins(prefix string) []string {
	var plugins []string
	pathDirs := filepath.SplitList(os.Getenv("PATH"))
	for _, dir := range pathDirs {
		files, _ := os.ReadDir(dir)
		for _, f := range files {
			name := f.Name()
			if strings.HasPrefix(name, prefix) && (f.Type().IsRegular() || f.Type()&os.ModeSymlink != 0) {
				plugins = append(plugins, name)
			}
		}
	}
	return plugins
}

func LookupInPath(prog string) (string, bool) {
	pathDirs := filepath.SplitList(os.Getenv("PATH"))
	for _, dir := range pathDirs {
		full := filepath.Join(dir, prog)
		if stat, err := os.Stat(full); err == nil && stat.Mode().IsRegular() && stat.Mode()&0111 != 0 {
			return full, true
		}
	}
	return "", false
}
