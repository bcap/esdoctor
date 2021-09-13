package client

import (
	"net/http"
)

func Mock(doRequest func(*http.Request, *http.Response) error) Versioned {
	fn := func(req *http.Request) (*http.Response, error) {
		resp := NewHttpResponse(req)
		return resp, doRequest(req, resp)
	}
	client, err := New("http://localhost:9200", WithRoundTripper(NewRoundTripper(fn)))
	if err != nil {
		panic("Mock should never get an error on New call with a mocked transport")
	}
	return client
}
