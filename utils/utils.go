package utils

import (
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"strings"
)

func DefaultKubeconfigPath() string {
	return filepath.Join(homedir.HomeDir(), ".kube", "config")
}

func SetConfigDefault() string {
	defaultConfigFile := os.Getenv("MC_CONFIG")
	if defaultConfigFile == "" {
		defaultConfigFile = "config.json" // valor padrão se a variável de ambiente não estiver definida
	}
	return defaultConfigFile
}

func ValidateConfig(cmd *cobra.Command, configFlagName, envVarName string) string {
	configFilePath, _ := cmd.Flags().GetString(configFlagName)

	if !cmd.Flags().Changed(configFlagName) {
		envConfigFile := os.Getenv(envVarName)
		if envConfigFile != "" {
			configFilePath = envConfigFile
		} else {
			configFilePath = "config.json"
		}
	}
	return configFilePath
}

func GetNamespaces(namespace string) []string {
	namespaces := strings.Split(namespace, ",")
	if len(namespaces) == 0 || (len(namespaces) == 1 && namespaces[0] == "") {
		return []string{""}
	}
	return namespaces
}

func getNodeStatus(node v1.Node) string {
	status := "Unknown"

	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady {
			if condition.Status == v1.ConditionTrue {
				status = "Ready"
			} else {
				status = "NotReady"
			}
			break
		}
	}

	if node.Spec.Unschedulable {
		if status == "Ready" {
			status = "Cordoned"
		} else if status == "NotReady" {
			status = "Unschedulable"
		}
	}

	return status
}

func getNodeRoles(node v1.Node) string {
	roles := []string{}

	for key := range node.Labels {
		if strings.HasPrefix(key, "node-role.kubernetes.io/") {
			roles = append(roles, strings.TrimPrefix(key, "node-role.kubernetes.io/"))
		}
	}

	if len(roles) == 0 {
		return "worker"
	}
	return strings.Join(roles, ",")
}
