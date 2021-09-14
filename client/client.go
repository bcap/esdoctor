package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	es5 "github.com/elastic/go-elasticsearch/v5"
	es6 "github.com/elastic/go-elasticsearch/v6"
	es7 "github.com/elastic/go-elasticsearch/v7"
	es8 "github.com/elastic/go-elasticsearch/v8"

	log "github.com/sirupsen/logrus"
)

type Versioned struct {
	V5       *es5.Client
	V6       *es6.Client
	V7       *es7.Client
	V8       *es8.Client
	endpoint string
}

type Option func(*config)

type config struct {
	endpoint               string
	roundTripper           http.RoundTripper
	logLevel               log.Level
	logRequestResponseBody bool
}

func WithEndpoint(endpoint string) Option {
	return func(config *config) {
		config.endpoint = endpoint
	}
}

func WithRoundTripper(roundTripper http.RoundTripper) Option {
	return func(config *config) {
		config.roundTripper = roundTripper
	}
}

func WithLogLevel(level log.Level) Option {
	return func(config *config) {
		config.logLevel = level
	}
}

func WithBodyLogging() Option {
	return func(config *config) {
		config.logRequestResponseBody = true
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

func newConfig(opts ...Option) config {
	result := config{
		logLevel: log.DebugLevel,
	}
	for _, fn := range opts {
		fn(&result)
	}
	return result
}

func New(endpoint string, options ...Option) (Versioned, error) {
	config := newConfig(options...)
	addresses := []string{sanitizeEndpoint(endpoint)}
	logger := NewRequestResponseLogger(
		config.logLevel,
		config.logRequestResponseBody,
		config.logRequestResponseBody,
	)

	errResult := func(majorVersion int, err error) (Versioned, error) {
		return Versioned{}, fmt.Errorf("failed to create client for Elasticsearch v%d: %w", majorVersion, err)
	}

	client5, err := es5.NewClient(es5.Config{
		Addresses: addresses,
		Transport: config.roundTripper,
		Logger:    logger,
	})
	if err != nil {
		return errResult(5, err)
	}

	client6, err := es6.NewClient(es6.Config{
		Addresses: addresses,
		Transport: config.roundTripper,
		Logger:    logger,
	})
	if err != nil {
		return errResult(6, err)
	}

	client7, err := es7.NewClient(es7.Config{
		Addresses: addresses,
		Transport: config.roundTripper,
		Logger:    logger,
	})
	if err != nil {
		return errResult(7, err)
	}

	client8, err := es8.NewClient(es8.Config{
		Addresses: addresses,
		Transport: config.roundTripper,
		Logger:    logger,
	})
	if err != nil {
		return errResult(8, err)
	}

	return Versioned{
		V5:       client5,
		V6:       client6,
		V7:       client7,
		V8:       client8,
		endpoint: addresses[0],
	}, nil
}

func (c Versioned) URL(path string) (*url.URL, error) {
	fullAddress := c.endpoint + "/" + sanitizePath(path)
	return url.Parse(fullAddress)
}

func (c Versioned) Request(ctx context.Context, method string, path string, headers http.Header, body io.ReadCloser) (*http.Response, error) {
	url, err := c.URL(path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, url.String(), body)
	if err != nil {
		return nil, err
	}
	if headers == nil {
		headers = make(http.Header)
	}
	req.Header = headers
	req.Body = body
	return c.Do(req)
}

func (c Versioned) Do(req *http.Request) (*http.Response, error) {
	return c.V8.Transport.Perform(req)
}

func (c Versioned) Endpoint() string {
	return c.endpoint
}

func sanitizeEndpoint(endpoint string) string {
	for len(endpoint) > 0 && endpoint[len(endpoint)-1] == '/' {
		endpoint = endpoint[:len(endpoint)-2]
	}
	return endpoint
}

func sanitizePath(path string) string {
	for len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	return path
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
