package stats

import (
	"context"

	"esdoctor/client"
	"esdoctor/fetch"
)

func GetNodes(ctx context.Context, client client.Versioned) (*Nodes, error) {
	result := Nodes{}
	return &result, fetch.Fetch(ctx, client, "_nodes/stats", &result)
}

func GetIndices(ctx context.Context, client client.Versioned) (*Indices, error) {
	result := Indices{}
	return &result, fetch.Fetch(ctx, client, "_stats?level=shards&expand_wildcards=all", &result)
}

func GetCluster(ctx context.Context, client client.Versioned) (*Cluster, error) {
	result := Cluster{}
	return &result, fetch.Fetch(ctx, client, "_cluster/stats", &result)
}

func GetTasks(ctx context.Context, client client.Versioned) (*Tasks, error) {
	result := Tasks{}
	return &result, fetch.Fetch(ctx, client, "_tasks?detailed=true&group_by=parents", &result)
}
