package stats

import (
	"context"

	"esdoctor/client"
	"esdoctor/fetch"
)

func GetNodesStats(ctx context.Context, client client.Versioned) (NodesStats, error) {
	result := NodesStats{}
	return result, fetch.Fetch(ctx, client, "_nodes/stats", &result)
}

func GetIndicesStats(ctx context.Context, client client.Versioned) (IndicesStats, error) {
	result := IndicesStats{}
	return result, fetch.Fetch(ctx, client, "_stats", &result)
}

func GetClusterStats(ctx context.Context, client client.Versioned) (ClusterStats, error) {
	result := ClusterStats{}
	return result, fetch.Fetch(ctx, client, "_cluster/stats", &result)
}
