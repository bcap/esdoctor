package diagnosis

import (
	"context"
	"fmt"

	"esdoctor/client"

	log "github.com/sirupsen/logrus"
)

func Diagnose(ctx context.Context, client client.Versioned, options ...Option) (*Diagnostics, error) {
	diagnostics := NewDiagnostics(client, options...)
	return diagnostics, diagnostics.Run(ctx)
}

func NewDiagnostics(client client.Versioned, options ...Option) *Diagnostics {
	return &Diagnostics{
		config: newConfig(options...),
		client: client,
	}
}

func (d *Diagnostics) Run(ctx context.Context) error {
	log.Infof("Running diagnostics on endpoint %s with the following config: %+v", d.client.Endpoint(), d.config)
	if err := d.load(ctx); err != nil {
		return fmt.Errorf("failed to load data for diagnistics: %w", err)
	}
	if err := d.process(); err != nil {
		return fmt.Errorf("failed to process loaded data: %w", err)
	}
	return nil
}

func (d *Diagnostics) process() error {
	return nil
}
