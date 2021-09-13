package client

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"

	es5 "github.com/elastic/go-elasticsearch/v5"
	es6 "github.com/elastic/go-elasticsearch/v6"
	es7 "github.com/elastic/go-elasticsearch/v7"
	es8 "github.com/elastic/go-elasticsearch/v8"
)

type Versioned struct {
	V5 *es5.Client
	V6 *es6.Client
	V7 *es7.Client
	V8 *es8.Client

	endpoint string
}

type Option func(*options)

func WithEndpoint(endpoint string) Option {
	return func(options *options) {
		addresses := []string{endpoint}
		options.es5Cfg.Addresses = addresses
		options.es6Cfg.Addresses = addresses
		options.es7Cfg.Addresses = addresses
		options.es8Cfg.Addresses = addresses
	}
}

func WithRoundTripper(transport http.RoundTripper) Option {
	return func(options *options) {
		options.es5Cfg.Transport = transport
		options.es6Cfg.Transport = transport
		options.es7Cfg.Transport = transport
		options.es8Cfg.Transport = transport
	}
}

// Simple delegate that implements http.RoundTripper
type roundTripper struct {
	do func(*http.Request) (*http.Response, error)
}

func (r roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return r.do(req)
}

func NewRoundTripper(fn func(*http.Request) (*http.Response, error)) http.RoundTripper {
	return roundTripper{do: fn}
}

type options struct {
	es5Cfg es5.Config
	es6Cfg es6.Config
	es7Cfg es7.Config
	es8Cfg es8.Config
}

func newOptions(opts ...Option) options {
	result := options{}
	for _, fn := range opts {
		fn(&result)
	}
	return result
}

func New(endpoint string, options ...Option) (Versioned, error) {
	opts := newOptions(options...)

	addresses := []string{endpoint}
	opts.es5Cfg.Addresses = addresses
	opts.es6Cfg.Addresses = addresses
	opts.es7Cfg.Addresses = addresses
	opts.es8Cfg.Addresses = addresses

	errResult := func(majorVersion int, err error) (Versioned, error) {
		return Versioned{}, fmt.Errorf("failed to create client for Elasticsearch v%d: %w", majorVersion, err)
	}

	client5, err := es5.NewClient(opts.es5Cfg)
	if err != nil {
		return errResult(5, err)
	}
	client6, err := es6.NewClient(opts.es6Cfg)
	if err != nil {
		return errResult(6, err)
	}
	client7, err := es7.NewClient(opts.es7Cfg)
	if err != nil {
		return errResult(7, err)
	}
	client8, err := es8.NewClient(opts.es8Cfg)
	if err != nil {
		return errResult(8, err)
	}

	return Versioned{
		V5:       client5,
		V6:       client6,
		V7:       client7,
		V8:       client8,
		endpoint: endpoint,
	}, nil
}

func (c Versioned) Endpoint() string {
	return c.endpoint
}

// Returns a basic, empty and succesfull http response
func NewHttpResponse(req *http.Request) *http.Response {
	data := "HTTP/1.1 200 OK\n\n"
	resp, err := http.ReadResponse(bufio.NewReader(strings.NewReader(data)), req)
	if err != nil {
		panic("NewHttpResponse should never fail. Root cause: " + err.Error())
	}
	return resp
}
