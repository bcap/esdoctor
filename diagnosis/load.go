package diagnosis

import (
	"context"
	"esdoctor/metadata"
	"esdoctor/stats"
	"esdoctor/version"

	log "github.com/sirupsen/logrus"
)

func (d *Diagnostics) load(ctx context.Context) error {
	log.Debug("Fetching supporting data")
	if err := d.loadVersion(ctx); err != nil {
		return err
	}

	var indicesMetadata metadata.IndicesMetadata
	// only load static information once
	if d.SupportingData.IndicesMetadata == nil {
		var err error
		indicesMetadata, err = metadata.GetIndexes(ctx, d.client)
		if err != nil {
			return err
		}
	} else {
		indicesMetadata = d.SupportingData.IndicesMetadata
	}

	indicesStats, err := stats.GetIndicesStats(ctx, d.client)
	if err != nil {
		return err
	}
	nodesStats, err := stats.GetNodesStats(ctx, d.client)
	if err != nil {
		return err
	}
	clusterStats, err := stats.GetClusterStats(ctx, d.client)
	if err != nil {
		return err
	}

	log.Info("Fetched supporting data")
	d.SupportingData = SupportingData{
		IndicesMetadata: indicesMetadata,
		IndicesStats:    indicesStats,
		NodesStats:      nodesStats,
		ClusterStats:    clusterStats,
	}
	return nil
}

func (d *Diagnostics) loadVersion(ctx context.Context) error {
	// only load static information once
	if d.Version.Set() {
		log.Debug("Elasticsearch version already discovered, not loading it")
		return nil
	}

	log.Debug("Discovering Elasticsearch version")
	version, err := version.Discover(ctx, d.client)
	if err != nil {
		return err
	}
	log.Infof("Elasticsearch is on version %s", version)
	d.Version = version
	return nil
}
