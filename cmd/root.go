package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "multicluster",
	Short: "CLI to manage multiple Kubernetes clusters",
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get Resources",
}

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe resources",
}

var (
	clusterName    string
	configFile     string
	kubeconfigPath string
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {

	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getPodsCmd)
	getCmd.AddCommand(getConfigMapsCmd)
	getCmd.AddCommand(getDeploymentsCmd)
	getCmd.AddCommand(getIngressCmd)
	getCmd.AddCommand(getLogsCmd)
	getCmd.AddCommand(getNamespacesCmd)
	getCmd.AddCommand(getNodesCmd)
	getCmd.AddCommand(getSecretsCmd)
	getCmd.AddCommand(getServicesCmd)

	rootCmd.AddCommand(describeCmd)
	describeCmd.AddCommand(describePodCmd)
}
