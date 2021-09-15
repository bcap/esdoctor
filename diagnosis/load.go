package diagnosis

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"esdoctor/hotthreads"
	"esdoctor/metadata"
	"esdoctor/stats"
	"esdoctor/version"

	log "github.com/sirupsen/logrus"
)

type dataCollection struct {
	version         version.ESVersion
	indicesMetadata metadata.Indices
	clusterState    *metadata.ClusterState
	clusterHealth   *metadata.ClusterHealth
	clusterStats    *stats.Cluster
	indicesStats    *stats.Indices
	nodesStats      *stats.Nodes
	tasks           *stats.Tasks
	hotThreads      *hotthreads.Group
}

func (d *Diagnostics) load(ctx context.Context) error {
	log.Debug("Fetching supporting data")

	var err error
	dc := dataCollection{}

	if dc.version, err = version.Discover(ctx, d.client); err != nil {
		return err
	}

	if dc.indicesMetadata, err = metadata.GetIndexes(ctx, d.client); err != nil {
		return err
	}

	if dc.clusterState, err = metadata.GetClusterState(ctx, d.client); err != nil {
		return err
	}

	if dc.clusterHealth, err = metadata.GetClusterHealth(ctx, d.client); err != nil {
		return err
	}

	if dc.indicesStats, err = stats.GetIndices(ctx, d.client); err != nil {
		return err
	}

	if dc.nodesStats, err = stats.GetNodes(ctx, d.client); err != nil {
		return err
	}

	if dc.clusterStats, err = stats.GetCluster(ctx, d.client); err != nil {
		return err
	}

	if dc.tasks, err = stats.GetTasks(ctx, d.client); err != nil {
		return err
	}

	if dc.hotThreads, err = hotthreads.Get(
		ctx, d.client,
		hotthreads.WithInterval(1*time.Second),
		hotthreads.WithTypes(hotthreads.TypeCPU, hotthreads.TypeCPU, hotthreads.TypeWait),
	); err != nil {
		return err
	}

	log.Info("Fetched supporting data")

	d.normalize(dc)

	return nil
}

func (d *Diagnostics) normalize(c dataCollection) {
	// version data normalization
	d.Version = c.version

	// hot thread data normalization
	d.HotThreads = c.hotThreads

	// cluster data normalization
	d.Cluster = &Cluster{
		State:  c.clusterState,
		Stats:  c.clusterStats,
		Health: c.clusterHealth,
	}

	// nodes data normalization
	d.Nodes = Nodes{
		Data:   map[string]*Node{},
		Master: map[string]*Node{},
		All:    map[string]*Node{},
	}
	for id, stats := range c.nodesStats.Nodes {
		entry := Node{ID: id, Name: stats.Name, Stats: stats}
		d.Nodes.All[id] = &entry
		for _, role := range stats.Roles {
			switch role {
			case "data":
				d.Nodes.Data[id] = &entry
			case "master":
				d.Nodes.Master[id] = &entry
			}
		}
	}

	// indices data normalization
	d.Indices = map[string]*Index{}
	for name, meta := range c.indicesMetadata {
		d.Indices[name] = &Index{
			Name:     name,
			Metadata: meta,
			Stats:    c.indicesStats.Indices[name],
		}
	}

	// shards data normalization + some index and node normalization due to shard locations
	for indexName, index := range c.clusterState.RoutingTable.Indices {
		normalizedIndex := d.Indices[indexName]
		indexStats := c.indicesStats.Indices[indexName]
		nodesWithIndexMap := map[string]struct{}{}
		nodesWithIndex := []*Node{}
		for shardID, shards := range index.Shards {
			shardsStats := indexStats.Shards[shardID]
			for _, shard := range shards {
				// find the stats for this shard
				var shardStats *stats.Shard
				for _, s := range shardsStats {
					if s.Routing.Node == shard.Node {
						shardStats = &s
					}
				}
				// create Shard entry
				normalizedNode := d.Nodes.Data[shard.Node]
				normalizedShard := Shard{
					ID:        shardID,
					IndexName: indexName,
					Index:     normalizedIndex,
					NodeID:    normalizedNode.ID,
					NodeName:  normalizedNode.Name,
					Node:      normalizedNode,
					State:     shard,
					Stats:     shardStats,
				}

				// create references for this shard in multiple places
				d.Shards = append(d.Shards, &normalizedShard)
				normalizedIndex.Shards = append(normalizedIndex.Shards, &normalizedShard)
				normalizedNode.Shards = append(normalizedNode.Shards, &normalizedShard)

				// mark this index being present in this node as it contains at least one shard on it
				if _, ok := nodesWithIndexMap[shard.Node]; !ok {
					nodesWithIndex = append(nodesWithIndex, normalizedNode)
					nodesWithIndexMap[shard.Node] = struct{}{}
				}
			}
		}
		sort.Slice(nodesWithIndex, func(i, j int) bool {
			return strings.Compare(nodesWithIndex[i].Name, nodesWithIndex[j].Name) == -1
		})
		normalizedIndex.Nodes = nodesWithIndex
	}

	// tasks data recursive normalization
	for _, task := range c.tasks.Tasks {
		d.Tasks = append(d.Tasks, d.normalizeTask(&task))
	}
}

func (d *Diagnostics) normalizeTask(task *stats.Task) *Task {
	result := Task{
		ID:       fmt.Sprintf("%s:%d", task.Node, task.ID),
		Task:     task,
		Node:     d.Nodes.All[task.Node],
		Children: []*Task{},
	}
	for _, child := range task.Children {
		normalizedChild := d.normalizeTask(&child)
		result.Children = append(result.Children, normalizedChild)
		normalizedChild.Parent = &result
	}
	return &result
}
