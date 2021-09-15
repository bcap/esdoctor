package hotthreads

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const nodePrefix = "::: "
const titlePrefix = "   Hot threads at "
const threadPrefix = "   "
const snapshotsPrefix = "     "
const stackLinePrefix = "       "

// from `toString` method in
//   https://github.com/elastic/elasticsearch/blob/master/server/src/main/java/org/elasticsearch/cluster/node/DiscoveryNode.java
// example string:
//   ::: {44bc264f76c34f8526d5464e499921b8}{jba9tWi1QVCQUXBjh-jbEw}{YxgG6NkiSTGAMhS7avYosA}{x.x.x.x}{x.x.x.x:9300}{mr}{dp_version=20210401, distributed_snapshot_deletion_enabled=false, cold_enabled=false, zone=us-east-1b, cross_cluster_transport_address=x.x.x.x, shard_indexing_pressure_enabled=true, }
var nodeIdPattern = regexp.MustCompile(`^\{([^\}]+)\}`)

// example string:
//   28.1% (140.5ms out of 500ms) cpu usage by thread 'elasticsearch[b60e77028320a7e79f0bd16a9c15cb61][refresh][T#4]'
// NOTE: Sometimes the last ' (single quote) character ending the thread name is not present. This is likely a bug in
//       the Amazon OpenSearch code as it sometimes hides the thread stack and name, replacing it with the
//       "[AMAZON INTERNAL]" string (but without the last ' char). Likely a bad string replacement filter
var threadPattern = regexp.MustCompile(`\d+\.\d+% \((\d+(?:\.\d+)?)(\w+) out of (\d+(?:\.\d+)?)(\w+)\) (\w+) usage by thread '([^']+)`)

// example string:
//   9/10 snapshots sharing following 8 elements
var snapshotsPattern = regexp.MustCompile(`(\d+)/\d+ snapshots sharing following \d+ elements`)

func Parse(data []byte) (*HotThreads, error) {
	lines := strings.Split(string(data), "\n")

	result := HotThreads{
		Nodes: map[string]*Node{},
	}
	var node *Node
	var thread *Thread
	var snapshotSummary *SnapshotSummary

	log.Tracef("Parsing %d lines of hot threads data", len(lines))

	for lineIdx, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// new node info
		if startsWith(line, nodePrefix) {
			matches := nodeIdPattern.FindStringSubmatch(line[4:])
			if matches == nil {
				return nil, fmt.Errorf("could not parse node line %d: %q", lineIdx+1, line)
			}
			node = &Node{
				ID:      matches[1],
				Threads: []*Thread{},
			}
			result.Nodes[node.ID] = node

		} else if startsWith(line, titlePrefix) {
			// TODO capture some info

		} else if startsWith(line, stackLinePrefix) {
			snapshotSummary.Stack = append(snapshotSummary.Stack, strings.TrimSpace(line))

		} else if startsWith(line, snapshotsPrefix+"unique snapshot") {
			snapshotSummary = &SnapshotSummary{
				Occurred: 1,
				Stack:    []string{},
			}
			thread.SnapshotSummaries = append(thread.SnapshotSummaries, snapshotSummary)

		} else if startsWith(line, snapshotsPrefix) {
			matches := snapshotsPattern.FindStringSubmatch(trimmed)
			if matches == nil {
				return nil, fmt.Errorf("could not parse snapshot line %d: %q", lineIdx+1, line)
			}
			occurred, _ := strconv.Atoi(matches[1])
			snapshotSummary = &SnapshotSummary{
				Occurred: occurred,
				Stack:    []string{},
			}
			thread.SnapshotSummaries = append(thread.SnapshotSummaries, snapshotSummary)

		} else if startsWith(line, threadPrefix) {
			matches := threadPattern.FindStringSubmatch(trimmed)
			if matches == nil {
				return nil, fmt.Errorf("could not parse thread line %d: %q", lineIdx+1, line)
			}

			taken, errTaken := strconv.ParseFloat(matches[1], 32)
			takenUnit, errTakenUnit := parseTimeUnit(matches[2])
			total, errTotal := strconv.ParseFloat(matches[3], 32)
			totalUnit, errTotalUnit := parseTimeUnit(matches[4])
			if errTaken != nil || errTotal != nil || errTakenUnit != nil || errTotalUnit != nil {
				return nil, fmt.Errorf("could not parse thread line %d: %q", lineIdx+1, line)
			}
			takenDuration := time.Duration(taken * float64(takenUnit))
			intervalDuration := time.Duration(total * float64(totalUnit))
			usagePercent := float64(takenDuration) / float64(intervalDuration) * 100.0

			usageType := matches[5]
			threadName := matches[6]

			// All threads should always be of the same collection type, so it should be ok to
			// set the type of the whole collection based on the first thread collection type we see
			if result.Type == "" {
				result.Type = CollectionType(usageType)
			}

			thread = &Thread{
				Type:              CollectionType(usageType),
				UsagePercent:      usagePercent,
				Time:              takenDuration,
				Interval:          intervalDuration,
				Name:              threadName,
				SnapshotSummaries: []*SnapshotSummary{},
			}
			node.Threads = append(node.Threads, thread)
		}
	}

	return &result, nil
}

func parseTimeUnit(unit string) (time.Duration, error) {
	switch unit {
	case "micros":
		return time.Microsecond, nil
	case "ms":
		return time.Millisecond, nil
	case "s":
		return time.Second, nil
	default:
		return time.ParseDuration(unit)
	}
}

func startsWith(s string, sub string) bool {
	if len(s) < len(sub) {
		return false
	}
	return s[0:len(sub)] == sub
}
