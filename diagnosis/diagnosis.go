package diagnosis

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"esdoctor/client"
	"esdoctor/math"
	"esdoctor/util"

	log "github.com/sirupsen/logrus"
)

func Diagnose(ctx context.Context, client client.Versioned, options ...Option) (*Diagnostics, error) {
	diagnostics := NewDiagnostics(client, options...)
	return diagnostics, diagnostics.Run(ctx)
}

func (d *Diagnostics) Run(ctx context.Context) error {
	log.Infof("Running diagnostics on endpoint %s with the following config: %+v", d.client.Endpoint(), d.config)

	if d.config.writer != nil {
		if err := d.config.writer.Begin(d); err != nil {
			return err
		}
	}

	if err := d.load(ctx); err != nil {
		return fmt.Errorf("failed to load data for diagnistics: %w", err)
	}

	if err := d.process(ctx); err != nil {
		return fmt.Errorf("failed to process loaded data: %w", err)
	}

	log.Info("Diagnosis is done")

	if d.config.writer != nil {
		if err := d.config.writer.End(d); err != nil {
			return err
		}
	}

	util.LogMemoryUsage("after fetching & generating all diagnostics data", log.DebugLevel)

	return nil
}

type DiagnosticsError struct {
	Errors []error
}

func (e *DiagnosticsError) Error() string {
	if len(e.Errors) == 0 {
		return "<no errors>"
	}
	return fmt.Sprintf("%d errors occurred while running diagnostics. Check logs", len(e.Errors))
}

func (d *Diagnostics) process(ctx context.Context) error {
	errors := []error{}
	for _, fn := range diagnosticsMethods {
		if err := fn(d, ctx); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return &DiagnosticsError{Errors: errors}
	}
	return nil
}

//
// Diagnosis methods
//

// TODO things to look for:
// - read and write rate per index. Avg latency per query, avg fetches per query
// - sharding vs num of nodes
// - balance between nodes
// - cluster colour, down to shard level
// - slowlog configs
// - refresh rate
// - flush rate
// - thread pools
// - master pending tasks
// - thread pool queue sizes
// - hot threads ?

var diagnosticsMethods = []func(*Diagnostics, context.Context) error{
	(*Diagnostics).processClusterHealth,
	(*Diagnostics).processReplicas,
	(*Diagnostics).processShardStates,
	(*Diagnostics).processNodesBalance,
	(*Diagnostics).processNodesDiskSizes,
	(*Diagnostics).processLuceneSegments,
}

const S001_ClusterGreen = "S001: " +
	"Cluster is in green status. All %d indices with a total of %d shards are available"

const W001_ClusterRed = "W001: " +
	"Cluster is in red status. The following indices have missing primary shards: %s"

const W002_ClusterYellow = "W002: " +
	"Cluster is in yellow status. The following indices have under-replicated shards: %s"

func (d *Diagnostics) processClusterHealth(ctx context.Context) error {
	colour := strings.ToLower(d.Cluster.Health.Status)
	if colour == "green" {
		d.Comment(S001_ClusterGreen, len(d.Indices), len(d.Shards))
		return nil
	}
	missingPrimaries := map[string]int{}
	missingReplicas := map[string]int{}
	for _, shard := range d.Shards {
		if shard.State.State == "STARTED" {
			continue
		}
		if shard.State.Primary {
			missingPrimaries[shard.IndexName]++
		} else {
			missingReplicas[shard.IndexName]++
		}
	}
	if colour == "red" {
		missingPrimariesMsg := []string{}
		for index, count := range missingPrimaries {
			missingPrimariesMsg = append(
				missingPrimariesMsg,
				fmt.Sprintf("%s (%d of %d)", index, count, len(d.Indices[index].Shards)),
			)
		}
		d.Comment(W001_ClusterRed, strings.Join(missingPrimariesMsg, ", "))
	} else if colour == "yellow" {
		missingReplicasMsg := []string{}
		for index, count := range missingReplicas {
			missingReplicasMsg = append(
				missingReplicasMsg,
				fmt.Sprintf("%s (%d of %d)", index, count, len(d.Indices[index].Shards)),
			)
		}
		d.Comment(W002_ClusterYellow, strings.Join(missingReplicasMsg, ", "))
	} else {
		return fmt.Errorf("Cluster is in unreconigzed status colour %q", d.Cluster.Health.Status)
	}
	return nil
}

const W003_NoReplicas = "W003: " +
	"Index %s has no replicas. In case a node that contains a shard of this index goes down, " +
	"the index will go automatically red and will need intervention (eg restore from a snapshot) " +
	"to be recovered. Searches will still work but will bring partial results (failed shards > 0 " +
	"in the response metadata), which normally clients are not aware. Currently this index is " +
	"present in %d of the %d data nodes (%.1f%% or %d/%d of the cluster)"

const A003_HighReplicas = "A003: " +
	"Index %s has %d replicas which is higher than 2. Normally a replication factor of 3x " +
	"(1 primary + 2 replicas) is enough to guarantee good enough resilience to node failures " +
	"and/or data loss. A high number of replicas may be desired though, particularly when you want " +
	"to improve search throughput, as multiple nodes can handle the search request. Currently this " +
	"index is present across %d out of %d data nodes (%.1f%% or %d/%d of the cluster)"

const I003_Replicas = "I003: " +
	"Index %s has %d replicas. Currently this index is present across %d out of %d data " +
	"nodes (%.1f%% or %d/%d of the cluster)"

const S003_Replicas = "S003: " +
	"%d indices out of %d (%.1f%%) have %d replicas"

func (d *Diagnostics) processReplicas(ctx context.Context) error {
	totalNodes := len(d.Nodes.Data)
	distribution := map[int]int{}
	for indexName, index := range d.Indices {
		numNodes := len(index.Nodes)
		denom, div, percentage := math.Fraction(int64(numNodes), int64(totalNodes))
		replicas, err := strconv.Atoi(index.Metadata.Settings.Index.NumberOfReplicas)
		if err != nil {
			log.Errorf("failed to read number of replicas for index %s: %v", indexName, err)
		} else if replicas == 0 {
			d.Comment(W003_NoReplicas, indexName, numNodes, totalNodes, percentage, denom, div)
		} else if replicas > 2 {
			d.Comment(A003_HighReplicas, indexName, replicas, numNodes, totalNodes, percentage, denom, div)
		} else {
			d.Comment(I003_Replicas, indexName, replicas, numNodes, totalNodes, percentage, denom, div)
		}
		distribution[replicas]++
	}

	for replicas, count := range distribution {
		d.Comment(S003_Replicas, count, len(d.Indices), math.Pct(count, len(d.Indices)), replicas)
	}

	return nil
}

const I004_ShardState = "I004: " +
	"%s shard %s of %s is in %s state and allocated in node %s. It contains %d documents " +
	"totalling %s (avg doc size of %s). It has %d lucene segments, with a " +
	"total memory utilization of %s"

const W004_ShardState = "W004: " +
	"%s shard %s of %s is in %s state"

const S004_ShardStates = "S004: " +
	"%d shards out of %d (%.1f%%) are in %s state"

func (d *Diagnostics) processShardStates(ctx context.Context) error {
	distribution := map[string]int{}
	for _, shard := range d.Shards {
		distribution[shard.State.State]++
		shardType := "primary"
		if !shard.State.Primary {
			shardType = "replica"
		}
		stats := shard.Stats
		avgDocSize := float64(stats.Store.SizeInBytes) / float64(stats.Docs.Count)
		d.Comment(
			I004_ShardState, shardType, shard.ID, shard.IndexName, shard.State.State,
			shard.NodeName, stats.Docs.Count, util.HumanizeBytes(stats.Store.SizeInBytes),
			util.HumanizeBytesF(avgDocSize), stats.Segments.Count,
			util.HumanizeBytes(int64(stats.Segments.MemoryInBytes)),
		)
		if shard.State.State != "STARTED" {
			d.Comment(W004_ShardState, shardType, shard.ID, shard.IndexName, shard.State.State)
		}
	}

	for state, count := range distribution {
		d.Comment(S004_ShardStates, count, len(d.Shards), math.Pct(count, len(d.Shards)), state)
	}

	return nil
}

const S005_NodeStorageDistribution = "S005: " +
	"The %d nodes have the following distribution in used disk space: " +
	"min=%s, p10=%s, p50=%s, p90=%s, max=%s"

const W005_NodeStorageUnbalanced = "W005: " +
	"Node %s storage seems unbalanced: its usage is %s %d%% of the current " +
	"median (50th percentile) of disk utilization: usage=%s, p50=%s. Check " +
	"the _cat/allocation, _cat/nodes and _cat/shards apis to better understand storage distribution. " +
	"Things to look for: large indices sharded in a small portion of the cluster (as opposed " +
	"to the whole cluster), nodes with different disk sizes (this tool should check for that as well) " +
	"or ongoing cluster replication/replacement of nodes"

// how much far off from the p50 we warn about inbalances in disk utilization
// TODO make it configurable
const unbalanceDiskUsageWarningFactor float64 = 0.2 // 20%

func (d *Diagnostics) processNodesBalance(ctx context.Context) error {
	distribution := []int64{}
	for _, node := range d.Nodes.Data {
		usage := node.Stats.Fs.Total.TotalInBytes - node.Stats.Fs.Total.AvailableInBytes
		distribution = append(distribution, usage)
	}
	pct := math.PercentilesInt64(distribution, 10)
	d.Comment(
		S005_NodeStorageDistribution, len(d.Nodes.Data), util.HumanizeBytes(pct[0]),
		util.HumanizeBytes(pct[1]), util.HumanizeBytes(pct[5]), util.HumanizeBytes(pct[9]),
		util.HumanizeBytes(pct[10]),
	)

	p50 := float64(pct[5])
	warningPercentage := int64(unbalanceDiskUsageWarningFactor * 100.0)
	aboveThreshold := p50 * (1.0 + unbalanceDiskUsageWarningFactor)
	belowThreshold := p50 * (1.0 - unbalanceDiskUsageWarningFactor)

	for _, node := range d.Nodes.Data {
		usage := float64(node.Stats.Fs.Total.TotalInBytes - node.Stats.Fs.Total.AvailableInBytes)
		if usage > aboveThreshold {
			d.Comment(
				W005_NodeStorageUnbalanced, node.Name, "above", warningPercentage,
				util.HumanizeBytesF(usage), util.HumanizeBytesF(p50),
			)
		} else if usage < belowThreshold {
			d.Comment(
				W005_NodeStorageUnbalanced, node.Name, "below", warningPercentage,
				util.HumanizeBytesF(usage), util.HumanizeBytesF(p50),
			)
		}
	}
	return nil
}

const W006_NodeStorageDifferentDiskSizes = "W006: " +
	"The cluster seems to have nodes with different amount of total disk space. " +
	"Current distribution: %v. This may indicate that data nodes have mixed hardware " +
	"profiles, which is non-ideal and will likely cause bottlenecking issues. You can use " +
	"the _cat/nodes, _cat/allocation or _nodes/stats apis to check for more details"

func (d *Diagnostics) processNodesDiskSizes(ctx context.Context) error {
	distribution := map[int64]int{}
	for _, node := range d.Nodes.Data {
		distribution[node.Stats.Fs.Total.TotalInBytes]++
	}
	if len(distribution) > 1 {
		distributionMsg := []string{}
		for size, numNodes := range distribution {
			distributionMsg = append(
				distributionMsg,
				fmt.Sprintf("%d nodes with %s", numNodes, util.HumanizeBytes(size)),
			)
		}
		d.Comment(W006_NodeStorageDifferentDiskSizes, strings.Join(distributionMsg, ", "))
	}
	return nil
}

const S006_LuceneSegmentsMemory = "S006: " +
	"Lucene segment memory utilization across the cluster is of %s, " +
	"distributed in the following: %s"

func (d *Diagnostics) processLuceneSegments(ctx context.Context) error {
	memoryDistribution := map[string]int64{}
	var memoryTotal int64
	for _, shard := range d.Shards {
		memoryTotal += int64(shard.Stats.Segments.MemoryInBytes)
		memoryDistribution["terms"] += int64(shard.Stats.Segments.TermsMemoryInBytes)
		memoryDistribution["stored_fields"] += int64(shard.Stats.Segments.StoredFieldsMemoryInBytes)
		memoryDistribution["term_vector"] += int64(shard.Stats.Segments.TermVectorsMemoryInBytes)
		memoryDistribution["norms"] += int64(shard.Stats.Segments.NormsMemoryInBytes)
		memoryDistribution["points"] += int64(shard.Stats.Segments.PointsMemoryInBytes)
		memoryDistribution["doc_values"] += int64(shard.Stats.Segments.DocValuesMemoryInBytes)
		memoryDistribution["index_writer"] += int64(shard.Stats.Segments.IndexWriterMemoryInBytes)
		memoryDistribution["version_map"] += int64(shard.Stats.Segments.VersionMapMemoryInBytes)
		memoryDistribution["fixed_bit_set"] += int64(shard.Stats.Segments.FixedBitSetMemoryInBytes)
	}

	keys := []string{}
	for key := range memoryDistribution {
		keys = append(keys, key)
	}
	// reverse sort from top to lower usage
	sort.Slice(keys, func(i int, j int) bool {
		return memoryDistribution[keys[i]] > memoryDistribution[keys[j]]
	})
	msg := []string{}
	for _, typ := range keys {
		numBytes := memoryDistribution[typ]
		msg = append(msg, fmt.Sprintf("%.1f%% %s", math.Pct64(numBytes, memoryTotal), typ))
	}
	d.Comment(S006_LuceneSegmentsMemory, util.HumanizeBytes(memoryTotal), strings.Join(msg, ", "))

	return nil
}
