package hotthreads

import (
	"time"
)

type CollectionType string

const TypeCPU CollectionType = "cpu"
const TypeBlock CollectionType = "block"
const TypeWait CollectionType = "wait"

type Group struct {
	CPU   *HotThreads `json:"cpu"`
	Block *HotThreads `json:"block"`
	Wait  *HotThreads `json:"wait"`
}

type HotThreads struct {
	Type  CollectionType   `json:"type"`
	Nodes map[string]*Node `json:"nodes"`
}

type Node struct {
	ID      string    `json:"id"`
	Threads []*Thread `json:"threads"`
}

type Thread struct {
	Name              string             `json:"thread"`
	UsagePercent      float64            `json:"usage_percent"`
	Interval          time.Duration      `json:"interval_ns"`
	Time              time.Duration      `json:"time_ns"`
	Type              CollectionType     `json:"type"`
	SnapshotSummaries []*SnapshotSummary `json:"snapshots"`
}

type SnapshotSummary struct {
	Occurred int      `json:"occurred"`
	Stack    []string `json:"stack"`
}

type NodeThreadPair struct {
	Node   *Node   `json:"node"`
	Thread *Thread `json:"thread"`
}

func (g *Group) set(typ CollectionType, ht *HotThreads) {
	switch typ {
	case TypeCPU:
		g.CPU = ht
	case TypeBlock:
		g.Block = ht
	case TypeWait:
		g.Wait = ht
	}
}
