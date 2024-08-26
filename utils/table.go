package utils

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"os"
	"strings"
)

func PrintNodeTable(clusterName string, nodes []v1.Node) {
	color.Cyan("Cluster: %s", clusterName)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Status", "Roles", "Age", "Version"})

	for _, node := range nodes {
		status := getNodeStatus(node)
		roles := getNodeRoles(node)
		//age := time.Since(node.CreationTimestamp.Time).Round(time.Second).String()
		age := node.CreationTimestamp.String()

		table.Append([]string{
			node.Name,
			status,
			roles,
			age,
			node.Status.NodeInfo.KubeletVersion,
		})
	}

	table.Render()
	fmt.Println()
}

func PrintDeploymentTable(clusterName string, deployments []appsv1.Deployment) {
	color.Cyan("Cluster: %s", clusterName)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Namespace", "Deployment Name", "Replicas", "Available Replicas", "Age"})

	for _, deployment := range deployments {
		table.Append([]string{
			deployment.Namespace,
			deployment.Name,
			fmt.Sprintf("%d", *deployment.Spec.Replicas),
			fmt.Sprintf("%d", deployment.Status.AvailableReplicas),
			deployment.CreationTimestamp.String(),
		})
	}

	table.Render()
	fmt.Println()
}

func PrintSecretTable(clusterName string, secrets []v1.Secret) {
	color.Cyan("Cluster: %s", clusterName)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Namespace", "Secret Name", "Type"})

	for _, secret := range secrets {
		table.Append([]string{secret.Namespace, secret.Name, string(secret.Type)})
	}

	table.Render()
	fmt.Println()
}

func PrintConfigMapTable(clusterName string, configMaps []v1.ConfigMap) {
	color.Cyan("Cluster: %s", clusterName)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Namespace", "ConfigMap Name"})

	for _, configMap := range configMaps {
		table.Append([]string{configMap.Namespace, configMap.Name})
	}

	table.Render()
	fmt.Println()
}

func PrintIngressTable(clusterName string, ingresses []networkingv1.Ingress) {
	color.Cyan("Cluster: %s", clusterName)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Cluster", "Namespace", "Ingress Name", "Hosts", "Paths", "Service"})

	for _, ing := range ingresses {
		hosts := []string{}
		paths := []string{}
		services := []string{}

		for _, rule := range ing.Spec.Rules {
			hosts = append(hosts, rule.Host)
			for _, path := range rule.HTTP.Paths {
				paths = append(paths, path.Path)
				if path.Backend.Service != nil {
					services = append(services, fmt.Sprintf("%s:%d", path.Backend.Service.Name, path.Backend.Service.Port.Number))
				}
			}
		}
		//address := ""
		//if len(ing.Status.LoadBalancer.Ingress) > 0 {
		//	address = ing.Status.LoadBalancer.Ingress[0].IP
		//}
		//
		//ports := ""
		//if ing.Spec.DefaultBackend != nil && ing.Spec.DefaultBackend.Service != nil {
		//	ports = fmt.Sprintf("%v", ing.Spec.DefaultBackend.Service.Port.Number)
		//}

		table.Append([]string{
			clusterName,
			ing.Namespace,
			ing.Name,
			fmt.Sprintf("%v", hosts),
			fmt.Sprintf("%v", paths),
			fmt.Sprintf("%v", services),
		})
	}

	table.Render()
	fmt.Println()
}

func PrintServiceTable(clusterName string, services []v1.Service) {
	color.Cyan("Cluster: %s", clusterName)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Namespace", "Service Name", "Type", "Cluster IP"})

	for _, service := range services {
		table.Append([]string{service.Namespace, service.Name, string(service.Spec.Type), service.Spec.ClusterIP})
	}

	table.Render()
	fmt.Println()
}

func PrintNamespaceTable(clusterName string, namespaces []v1.Namespace) {
	color.Cyan("Cluster: %s", clusterName)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Namespace"})

	for _, namespace := range namespaces {
		table.Append([]string{namespace.Name})
	}

	table.Render()
	fmt.Println()
}

func PrintPodTable(clusterName string, pods []v1.Pod, statusFilter string) {
	color.Cyan("Cluster: %s", clusterName)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Namespace", "Pod Name", "Ready", "Status", "Restarts", "Node"})

	for _, pod := range pods {
		if statusFilter != "" && strings.ToLower(string(pod.Status.Phase)) != strings.ToLower(statusFilter) {
			continue
		}

		readyContainers := 0
		totalRestarts := int32(0)
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Ready {
				readyContainers++
			}
			totalRestarts += containerStatus.RestartCount
		}
		containersReady := fmt.Sprintf("%d/%d", readyContainers, len(pod.Status.ContainerStatuses))
		restarts := fmt.Sprintf("%d", totalRestarts)

		table.Append([]string{pod.Namespace, pod.Name, containersReady, string(pod.Status.Phase), restarts, pod.Spec.NodeName})
	}

	table.Render()
	fmt.Println()
}

func PrintPodLogs(clusterName, namespace, podName string, logs io.ReadCloser, follow bool) {
	color.Cyan("Cluster: %s", clusterName)
	color.Cyan("Namespace: %s", namespace)
	color.Cyan("Pod: %s", podName)
	fmt.Println()

	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading log stream:", err)
	}
}
