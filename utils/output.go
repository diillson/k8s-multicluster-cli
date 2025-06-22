package utils

import (
	"encoding/json"
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/handlers"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"time"
)

func OutputDelete(results []handlers.DeleteResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: ERROR - %s\n", res.Cluster.Name, res.Err)
		} else if res.Success {
			fmt.Printf("Cluster %s: Recurso deletado com sucesso! %s\n", res.Cluster.Name, res.Message)
		} else {
			fmt.Printf("Cluster %s: Falha ao deletar recurso. %s\n", res.Cluster.Name, res.Message)
		}
	}
}

func OutputApply(results []handlers.ApplyResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s | %-12s/%-12s %-20s : ERROR: %v\n", res.Cluster.Name, res.Namespace, res.Resource, res.Name, res.Err)
		} else if res.Success {
			fmt.Printf("Cluster %s | %-12s/%-12s %-20s : APPLIED OK\n", res.Cluster.Name, res.Namespace, res.Resource, res.Name)
		} else {
			fmt.Printf("Cluster %s | %-12s/%-12s %-20s : [N/A]\n", res.Cluster.Name, res.Namespace, res.Resource, res.Name)
		}
	}
}

func OutputPatch(results []handlers.PatchResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s | %s/%s: PATCH ERROR: %v\n", res.Cluster.Name, res.Namespace, res.Name, res.Err)
		} else if res.Success {
			fmt.Printf("Cluster %s | %s/%s: PATCH SUCESSO\n", res.Cluster.Name, res.Namespace, res.Name)
		} else {
			fmt.Printf("Cluster %s | %s/%s: [UNKNOWN/NO PATCH]\n", res.Cluster.Name, res.Namespace, res.Name)
		}
	}
}

func OutputDescribe(results []handlers.DescribeResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: ERRO: %v\n", res.Cluster.Name, res.Err)
			continue
		}
		fmt.Printf("Cluster: %s | Namespace: %s | Resource: %s/%s\n", res.Cluster.Name, res.Namespace, res.Object.GetKind(), res.Name)
		printUnstructuredObject(res.Object)
		fmt.Println("-----\n")
	}
}

// Pode imprimir como YAML/JSON bonito, ou então detalhar os campos.
func printUnstructuredObject(obj *unstructured.Unstructured) {
	if obj == nil {
		fmt.Println("Objeto não encontrado.")
		return
	}
	by, _ := json.MarshalIndent(obj.Object, "", "  ")
	fmt.Println(string(by))
}

func OutputLogs(results []handlers.LogsResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s | %s/%s: [ERR] %v\n", res.Cluster.Name, res.Namespace, res.PodName, res.Err)
			continue
		}
		fmt.Printf("\n--- Cluster: %s | Namespace: %s | Pod: %s | Container: %s ---\n", res.Cluster.Name, res.Namespace, res.PodName, res.Container)
		fmt.Print(res.Logs)
	}
}

func OutputExec(results []handlers.ExecResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s | %s/%s: [ERR] %v\n", res.Cluster.Name, res.Namespace, res.PodName, res.Err)
			continue
		}
		fmt.Printf("\n--- Cluster: %s | Namespace: %s | Pod: %s | Container: %s ---\n", res.Cluster.Name, res.Namespace, res.PodName, res.Container)
		fmt.Print(res.Output)
	}
}

func OutputAPIResources(results []handlers.APIResourcesResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: ERRO: %v\n", res.Cluster.Name, res.Err)
			continue
		}
		fmt.Printf("Cluster %s:\n", res.Cluster.Name)
		for _, resource := range res.Resources {
			fmt.Println("- " + resource)
		}
		fmt.Println("------")
	}
}

func OutputAPIVersions(results []handlers.APIVersionsResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: ERRO: %v\n", res.Cluster.Name, res.Err)
			continue
		}
		fmt.Printf("Cluster %s:\n", res.Cluster.Name)
		for _, v := range res.Versions {
			fmt.Println("- " + v)
		}
		fmt.Println("------")
	}
}

func OutputEvents(results []handlers.EventsResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s | Namespace %s: ERRO ao buscar eventos: %v\n", res.Cluster.Name, res.Namespace, res.Err)
			continue
		}
		fmt.Printf("\nCluster: %s | Namespace: %s\n", res.Cluster.Name, res.Namespace)
		fmt.Printf("%-30s %-20s %-10s %-30s %-20s\n", "LAST SEEN", "TYPE", "REASON", "OBJECT", "MESSAGE")
		for _, ev := range res.Events {
			fmt.Printf("%-30s %-20s %-10s %-30s %-20s\n",
				timeAgo(ev.LastTimestamp.Time),
				ev.Type,
				ev.Reason,
				ev.InvolvedObject.Kind+"/"+ev.InvolvedObject.Name,
				ev.Message)
		}
		fmt.Println("--------------------------------")
	}
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d.Hours() > 24:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d.Hours() > 1:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d.Minutes() > 1:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	default:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
}

func OutputVersion(results []handlers.VersionResult) {
	fmt.Println("\n---- SERVIDORES ----")
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: ERRO ao buscar versão do servidor: %v\n", res.Cluster.Name, res.Err)
			continue
		}
		fmt.Printf("Cluster: %-17s Version: %-12s Platform: %s\n",
			res.Cluster.Name,
			res.GitVersion,
			res.Platform,
		)
	}
}

func OutputPortForward(results []handlers.PortForwardResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: Falhou port-forward %s/%s: %v\n", res.Cluster.Name, res.Namespace, res.Name, res.Err)
		} else if res.Success {
			fmt.Printf("Cluster %s: Port-forward iniciado para %s/%s: %v@%s\n", res.Cluster.Name, res.Namespace, res.Name, res.Ports, res.Address)
		}
	}
}

func OutputWhoami(results []handlers.WhoamiResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: ERRO: %v\n", res.Cluster.Name, res.Err)
			continue
		}
		fmt.Printf("\nCluster: %s\n", res.Cluster.Name)
		fmt.Printf("Usuario:  %s\n", res.Username)
		fmt.Printf("UID:      %s\n", res.UID)
		fmt.Printf("Grupos:   %v\n", res.Groups)
		for k, v := range res.Extras {
			fmt.Printf("Extra[%s]: %v\n", k, v)
		}
		fmt.Println("------")
	}
}

func OutputCanI(results []handlers.CanIResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: ERRO: %v\n", res.Cluster.Name, res.Err)
			continue
		}
		status := "NO"
		if res.Allowed {
			status = "YES"
		}
		fmt.Printf("Cluster: %-16s | Can %s %s/%s: %-3s [%s]\n",
			res.Cluster.Name, res.Verb, res.Resource, res.Name, status, res.Reason)
	}
}

func OutputDrain(results []handlers.DrainResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: Falha drain no node %s: %v (%s)\n", res.Cluster.Name, res.Node, res.Err, res.Message)
		} else if res.Success {
			fmt.Printf("Cluster %s: Node %s drenado!\n", res.Cluster.Name, res.Node)
		} else {
			fmt.Printf("Cluster %s: DRAIN parcial/falhou em %s: %s\n", res.Cluster.Name, res.Node, res.Message)
		}
	}
}

func OutputCordon(results []handlers.CordonResult, mode string) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: Falha ao %s node %s: %v (%s)\n", res.Cluster.Name, mode, res.Node, res.Err, res.Message)
		} else if res.Success {
			fmt.Printf("Cluster %s: Node %s %sed.\n", res.Cluster.Name, res.Node, mode)
		} else {
			fmt.Printf("Cluster %s: %s parcial/falhou: %s\n", res.Cluster.Name, mode, res.Message)
		}
	}
}

func OutputLabel(results []handlers.LabelResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: Falha label: %v (%s)\n", res.Cluster.Name, res.Err, res.Message)
		} else {
			fmt.Printf("Cluster %s: Label definida com sucesso\n", res.Cluster.Name)
		}
	}
}
func OutputAnnotate(results []handlers.AnnotateResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: Falha annotate: %v (%s)\n", res.Cluster.Name, res.Err, res.Message)
		} else {
			fmt.Printf("Cluster %s: Annotation definida com sucesso\n", res.Cluster.Name)
		}
	}
}

func OutputScale(results []handlers.ScaleResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: Falhou scale %s/%s: %v (%s)\n", res.Cluster.Name, res.Namespace, res.Name, res.Err, res.Message)
		} else if res.Success {
			fmt.Printf("Cluster %s: %s/%s escalado para %d\n", res.Cluster.Name, res.Namespace, res.Name, res.Replicas)
		} else {
			fmt.Printf("Cluster %s: SCALE parcial/falhou: %s\n", res.Cluster.Name, res.Message)
		}
	}
}

func OutputRolloutStatus(results []handlers.RolloutStatusResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: rollout falhou %s/%s: %v (%s)\n", res.Cluster.Name, res.Namespace, res.Name, res.Err, res.Message)
		} else if res.Done {
			fmt.Printf("Cluster %s: rollout completado %s/%s\n", res.Cluster.Name, res.Namespace, res.Name)
		} else {
			fmt.Printf("Cluster %s: rollout pendente %s/%s: %s\n", res.Cluster.Name, res.Namespace, res.Name, res.Message)
		}
	}
}

func OutputRolloutRestart(results []handlers.RolloutRestartResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: rollout restart falhou %s/%s: %v (%s)\n", res.Cluster.Name, res.Namespace, res.Name, res.Err, res.Message)
		} else if res.Success {
			fmt.Printf("Cluster %s: rollout restart disparado em %s/%s\n", res.Cluster.Name, res.Namespace, res.Name)
		}
	}
}

func OutputEdit(results []handlers.EditUpdateResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: Falha ao editar %s/%s: %v\n", res.Cluster.Name, res.Namespace, res.Name, res.Err)
		} else if res.Success {
			fmt.Printf("Cluster %s: %s/%s atualizado!\n", res.Cluster.Name, res.Namespace, res.Name)
		}
	}
}

func OutputSetLastApplied(results []handlers.SetLastAppliedResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: Falhou ao salvar last-applied: %v\n", res.Cluster.Name, res.Err)
		} else if res.Success {
			fmt.Printf("Cluster %s: last-applied salvo!\n", res.Cluster.Name)
		}
	}
}

func OutputViewLastApplied(results []handlers.ViewLastAppliedResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: Falhou ao buscar last-applied: %v\n", res.Cluster.Name, res.Err)
		} else if res.LastApplied == "" {
			fmt.Printf("Cluster %s: Recurso não tem annotation.\n", res.Cluster.Name)
		} else {
			fmt.Printf("Cluster %s: Last-applied configuration:\n%s\n", res.Cluster.Name, res.LastApplied)
		}
	}
}

func OutputRolloutHistory(results []handlers.RolloutHistoryResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: history falhou %s/%s: %v\n", res.Cluster.Name, res.Namespace, res.Name, res.Err)
			continue
		}
		fmt.Printf("Cluster %s: history de %s/%s:\n", res.Cluster.Name, res.Namespace, res.Name)
		for _, r := range res.Revisions {
			fmt.Printf("- %s\n", r)
		}
	}
}

func OutputRolloutUndo(results []handlers.RolloutUndoResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: rollout undo falhou %s/%s: %v\n", res.Cluster.Name, res.Namespace, res.Name, res.Err)
		} else if res.Success {
			fmt.Printf("Cluster %s: undo para %s/%s revisao #%d completo\n", res.Cluster.Name, res.Namespace, res.Name, res.Revision)
		}
	}
}

func OutputCreateCronJob(results []handlers.CreateCronJobResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: CronJob %s/%s falhou: %v\n", res.Cluster.Name, res.Namespace, res.Name, res.Err)
		} else {
			fmt.Printf("Cluster %s: CronJob %s/%s criado!\n", res.Cluster.Name, res.Namespace, res.Name)
		}
	}
}

func OutputCreateJob(results []handlers.CreateJobResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: Job %s/%s falhou: %v\n", res.Cluster.Name, res.Namespace, res.Name, res.Err)
		} else {
			fmt.Printf("Cluster %s: Job %s/%s criado!\n", res.Cluster.Name, res.Namespace, res.Name)
		}
	}
}

func OutputLogsJob(results []handlers.LogsJobResult) {
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Cluster %s: pod=%s ERR=%v\n", res.Cluster.Name, res.PodName, res.Err)
		} else {
			fmt.Printf("\nCluster: %s | Pod: %s\n", res.Cluster.Name, res.PodName)
			fmt.Print(res.Stdout)
		}
	}
}
