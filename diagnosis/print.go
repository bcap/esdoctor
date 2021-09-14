package diagnosis

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/davecgh/go-spew/spew"
	"github.com/imdario/mergo"
)

const FormatDumpJSON = "json-dump"
const FormatDumpSpew = "spew-dump"

type PrintOption func(*printConfig)

func WithPrintFormat(format string) PrintOption {
	return func(config *printConfig) {
		config.format = format
	}
}

type printConfig struct {
	format string
}

func newPrintConfig(options ...PrintOption) printConfig {
	config := printConfig{
		format: FormatDumpJSON,
	}
	for _, fn := range options {
		fn(&config)
	}
	return config
}

func (d Diagnostics) Print(writer io.Writer, options ...PrintOption) error {
	config := newPrintConfig(options...)
	switch config.format {
	case FormatDumpSpew:
		d.SpewDump(writer)
		return nil
	case FormatDumpJSON:
		return d.JSONDump(writer)
	default:
		return fmt.Errorf("unrecognized print format %q", config.format)
	}
}

func (d Diagnostics) SpewDump(writer io.Writer) {
	// we want to spew dump only the public members of the struct
	result := Diagnostics{}
	mergo.Merge(result, d)
	spew.Fdump(writer, result)
}

func (d Diagnostics) JSONDump(writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(d)
}
