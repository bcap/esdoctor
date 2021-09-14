package diagnosis

import (
	"esdoctor/client"
	"esdoctor/metadata"
	"esdoctor/stats"
	"esdoctor/version"
)

type Option func(*config)

type config struct {
}

func newConfig(optionFns ...Option) config {
	config := config{}
	for _, fn := range optionFns {
		fn(&config)
	}
	return config
}

type Diagnostics struct {
	Version        version.ESVersion
	SupportingData SupportingData

	client client.Versioned
	config config
}

type SupportingData struct {
	IndicesMetadata metadata.IndicesMetadata
	IndicesStats    stats.IndicesStats
	NodesStats      stats.NodesStats
	ClusterStats    stats.ClusterStats
}
