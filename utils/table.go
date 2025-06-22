package utils

import (
	"encoding/json"
	"fmt"
	"github.com/diillson/k8s-multicluster-cli/models"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"strings"
	"time"
)

func OutputGeneric(results []models.GenericResourceListResult, format string, resourceType string) {
	switch format {
	case "json":
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
		return
	case "yaml":
		data, _ := yaml.Marshal(results)
		fmt.Println(string(data))
		return
	}
	// Table (default)
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("Erro no cluster %s: %v\n", res.Cluster.Name, res.Err)
			continue
		}
		PrintUnstructuredTable(res.Cluster.Name, resourceType, res.Items)
	}
}

// Para printar recursos arbitrários:
func PrintUnstructuredTable(cluster, resourceType string, items []unstructured.Unstructured) {
	fmt.Printf("\nCluster: %s — Resource: %s\n", cluster, resourceType)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAMESPACE", "NAME", "AGE"})
	for _, item := range items {
		meta := item.Object
		ns := extractString(meta, "metadata.namespace")
		name := extractString(meta, "metadata.name")
		age := extractAge(meta)
		table.Append([]string{ns, name, age})
	}
	table.Render()
}

// --- Helpers já mostrados antes:

func extractString(obj map[string]interface{}, path string) string {
	keys := strings.Split(path, ".")
	cur := obj
	for i, key := range keys {
		if i == len(keys)-1 {
			v, _ := cur[key]
			if val, ok := v.(string); ok {
				return val
			}
		}
		v, _ := cur[key]
		if next, ok := v.(map[string]interface{}); ok {
			cur = next
		}
	}
	return ""
}

func extractAge(obj map[string]interface{}) string {
	m := extractString(obj, "metadata.creationTimestamp")
	if m == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, m)
	if err != nil {
		return m
	}
	return fmt.Sprintf("%dd", int(time.Since(t).Hours()/24))
}
