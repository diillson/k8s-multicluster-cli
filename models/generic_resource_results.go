package models

import (
	"github.com/diillson/k8s-multicluster-cli/internal/executor"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type GenericResourceListResult struct {
	Cluster executor.ClusterInfo
	Items   []unstructured.Unstructured
	Err     error
}
