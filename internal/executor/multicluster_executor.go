package executor

import (
	"context"
	"sync"
)

type ClusterInfo struct {
	Name    string
	Context string
}

type MultiClusterHandler interface {
	RunOnCluster(ctx context.Context, cluster ClusterInfo, kubeConfigPath string) error
	Result() interface{}
}

type ExecutionResult struct {
	Cluster ClusterInfo
	Err     error
}

func ExecuteOnClusters(
	ctx context.Context,
	handler MultiClusterHandler,
	clusters []ClusterInfo,
	kubeconfigPath string,
	maxWorkers int,
) []ExecutionResult {
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxWorkers)
	results := make([]ExecutionResult, len(clusters))

	for i, cluster := range clusters {
		wg.Add(1)
		go func(i int, cluster ClusterInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			err := handler.RunOnCluster(ctx, cluster, kubeconfigPath)
			results[i] = ExecutionResult{
				Cluster: cluster,
				Err:     err,
			}
		}(i, cluster)
	}
	wg.Wait()
	return results
}
