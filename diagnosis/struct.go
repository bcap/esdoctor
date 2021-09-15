package diagnosis

import (
	"encoding/json"
	"io"
	"os"
	"sync"

	"esdoctor/client"
	"esdoctor/hotthreads"
	"esdoctor/metadata"
	"esdoctor/stats"
	"esdoctor/version"

	"github.com/davecgh/go-spew/spew"
	"github.com/imdario/mergo"
)

type Option func(*config)

func WithOutput(w CommentWriter) Option {
	return func(c *config) {
		c.writer = w
	}
}

type config struct {
	writer CommentWriter
}

func newConfig(optionFns ...Option) config {
	config := config{
		writer: NewTextCommentWriter(os.Stdout, nil, false),
	}
	for _, fn := range optionFns {
		fn(&config)
	}
	return config
}

type Diagnostics struct {
	Version    version.ESVersion `json:"version"`
	Cluster    *Cluster          `json:"cluster"`
	Nodes      Nodes             `json:"nodes"`
	Indices    map[string]*Index `json:"indices"`
	Shards     []*Shard          `json:"shards"`
	Tasks      []*Task           `json:"tasks"`
	HotThreads *hotthreads.Group `json:"hot_threads"`

	comments    []Comment
	commentLock sync.RWMutex

	client client.Versioned
	config config
}

type Cluster struct {
	State  *metadata.ClusterState  `json:"state"`
	Stats  *stats.Cluster          `json:"stats"`
	Health *metadata.ClusterHealth `json:"health"`
}

type Nodes struct {
	Data   map[string]*Node `json:"data"`
	Master map[string]*Node `json:"master"`
	All    map[string]*Node `json:"all"`
}

type Node struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Stats  *stats.Node `json:"stats"`
	Shards []*Shard    `json:"shards"`
}

type Shard struct {
	ID        string               `json:"id"`
	IndexName string               `json:"index"`
	Index     *Index               `json:"-"` // backlink, avoid cyclic serialization
	NodeID    string               `json:"node_id"`
	NodeName  string               `json:"node_name"`
	Node      *Node                `json:"-"` // backlink, avoid cyclic serialization
	State     *metadata.ShardState `json:"state"`
	Stats     *stats.Shard         `json:"stats"`
}

type Index struct {
	Name     string          `json:"name"`
	Stats    *stats.Index    `json:"stats"`
	Metadata *metadata.Index `json:"metadata"`
	Nodes    []*Node         `json:"-"` // backlink, avoid cyclic serialization
	Shards   []*Shard        `json:"shards"`
}

type Task struct {
	*stats.Task
	ID       string  `json:"canonical_id"`
	Node     *Node   `json:"-"` // expands to excessive data on serialiation
	Parent   *Task   `json:"-"` // backlink, avoid cyclic serialization
	Children []*Task `json:"children"`
}

func NewDiagnostics(client client.Versioned, options ...Option) *Diagnostics {
	return &Diagnostics{
		config: newConfig(options...),
		client: client,
	}
}

func (d *Diagnostics) SpewDump(writer io.Writer) {
	// we want to spew dump only the public members of the struct, so we use mergo.Merge to
	// copy over the public members into an empty struct
	result := Diagnostics{}
	mergo.Merge(&result, d)
	spew.Fdump(writer, &result)
}

func (d *Diagnostics) JSONDump(writer io.Writer) error {
	wrapped := struct {
		*Diagnostics
		Comments []Comment `json:"comments"`
	}{
		Diagnostics: d,
		Comments:    d.Comments(),
	}
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(wrapped)
}
