package diagnosis

import (
	"context"
	"esdoctor/client"
	"esdoctor/version"
	"io"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

type Option func(*options)

type options struct {
}

func newOptions(optionFns ...Option) options {
	options := options{}
	for _, fn := range optionFns {
		fn(&options)
	}
	return options
}

func Diagnose(ctx context.Context, client client.Versioned, optionFns ...Option) (Diagnostics, error) {
	options := newOptions(optionFns...)
	log.Debugf("Running diagnostics on endpoint %s with the following options: %+v", client.Endpoint(), options)

	version, err := version.Discover(ctx, client)
	if err != nil {
		return Diagnostics{}, err
	}

	return Diagnostics{
		Version: version,
	}, nil
}

type Diagnostics struct {
	Version version.ESVersion
}

func (d Diagnostics) Print(writer io.Writer) {
	spew.Fprintf(writer, "%+v\n", d)
}
