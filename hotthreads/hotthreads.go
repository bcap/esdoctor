package hotthreads

import (
	"context"
	"fmt"
	"io/ioutil"
	"sort"
	"time"

	"esdoctor/client"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// ES hot threads:
// - docs: https://www.elastic.co/guide/en/elasticsearch/reference/master/cluster-nodes-hot-threads.html
// - code:
//   - https://github.com/elastic/elasticsearch/blob/master/server/src/main/java/org/elasticsearch/rest/action/admin/cluster/RestNodesHotThreadsAction.java
//   - https://github.com/elastic/elasticsearch/blob/master/server/src/main/java/org/elasticsearch/monitor/jvm/HotThreads.java

type Option func(*config)

func WithInterval(interval time.Duration) Option {
	return func(c *config) {
		c.interval = interval
	}
}

func WithSnapshots(snapshots int) Option {
	return func(c *config) {
		c.snapshots = snapshots
	}
}

func WithThreads(threads int) Option {
	return func(c *config) {
		c.threads = threads
	}
}

func WithTypes(types ...CollectionType) Option {
	return func(c *config) {
		c.types = types
	}
}

func WithAddType(typ CollectionType) Option {
	return func(c *config) {
		for _, t := range c.types {
			if t == typ {
				return
			}
		}
		c.types = append(c.types, typ)
	}
}

type config struct {
	interval  time.Duration
	snapshots int
	threads   int
	types     []CollectionType
}

func newConfig(options ...Option) config {
	config := config{
		interval:  500 * time.Millisecond,
		snapshots: 10,
		threads:   10,
		types:     []CollectionType{TypeCPU},
	}
	for _, fn := range options {
		fn(&config)
	}
	return config
}

func Get(ctx context.Context, client client.Versioned, options ...Option) (*Group, error) {
	config := newConfig(options...)
	log.Debugf("Fetching hot threads with config: %+v", config)

	group := Group{}
	executor, ctx := errgroup.WithContext(ctx)
	for _, t := range config.types {
		typ := t
		executor.Go(func() error {
			ht, err := getSingle(ctx, client, typ, config)
			if err != nil {
				return err
			}
			group.set(typ, ht)
			return nil
		})
	}
	return &group, executor.Wait()
}

func getSingle(ctx context.Context, client client.Versioned, collectionType CollectionType, config config) (*HotThreads, error) {
	api := fmt.Sprintf(
		"_nodes/hot_threads?interval=%v&snapshots=%d&threads=%d&type=%s",
		config.interval, config.snapshots, config.threads, collectionType,
	)
	resp, err := client.Request(ctx, "GET", api, nil, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch %s, got status code %d from ES", api, resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", api, err)
	}

	result, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse body from %s: %w", api, err)
	}

	return result, nil
}

// Convenience function to sort threads from multiple collections across the cluster,
// from coldest to hottest
func SortByUsage(htts ...*HotThreads) []NodeThreadPair {
	return Sort(func(a *NodeThreadPair, b *NodeThreadPair) bool {
		return a.Thread.UsagePercent < b.Thread.UsagePercent
	}, htts...)
}

func Sort(less func(a *NodeThreadPair, b *NodeThreadPair) bool, htts ...*HotThreads) []NodeThreadPair {
	result := []NodeThreadPair{}
	for _, h := range htts {
		for _, node := range h.Nodes {
			for _, thread := range node.Threads {
				result = append(result, NodeThreadPair{Node: node, Thread: thread})
			}
		}
	}
	sort.Slice(result, func(i int, j int) bool {
		return less(&result[i], &result[j])
	})
	return result
}
