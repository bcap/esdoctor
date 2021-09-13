package version

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"esdoctor/client"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	version, err := Parse("7.13.3")
	assert.NoError(t, err)
	assert.Equal(t, ESVersion{Major: 7, Minor: 13, Patch: 3}, version)

	mustFailInpus := []string{
		"foo", "foo.bar", "foo.bar.baz",
		"7", "7.13",
		"7.13.3.7",
		"7.bar.3",
		"7.13.3_alpha",
	}
	for _, input := range mustFailInpus {
		_, err := Parse(input)
		assert.Error(t, err)
	}
}

func TestDiscover5(t *testing.T) {
	testDiscover(
		t, ESVersion{Major: 5, Minor: 6, Patch: 17},
		` {
			"name": "wG4i22U",
			"cluster_name": "053131491888:ds-10-cluster10",
			"cluster_uuid": "xIc_5CypSZiRGZw5FvL3QQ",
			"version": {
			  "number": "5.6.17",
			  "build_hash": "59bb0dc",
			  "build_date": "2020-01-03T11:28:23.851Z",
			  "build_snapshot": false,
			  "lucene_version": "6.6.1"
			},
			"tagline": "You Know, for Search"
		  }`,
	)
}

func TestDiscover7(t *testing.T) {
	testDiscover(
		t, ESVersion{Major: 7, Minor: 13, Patch: 3},
		` {
			"name": "node-0",
			"cluster_name": "elasticsearch-dev",
			"cluster_uuid": "7hOdSLhiQ4W2IpXOU82HAg",
			"version": {
			  "number": "7.13.3",
			  "build_flavor": "default",
			  "build_type": "tar",
			  "build_hash": "5d21bea28db1e89ecc1f66311ebdec9dc3aa7d64",
			  "build_date": "2021-07-02T12:06:10.804015202Z",
			  "build_snapshot": false,
			  "lucene_version": "8.8.2",
			  "minimum_wire_compatibility_version": "6.8.0",
			  "minimum_index_compatibility_version": "6.0.0-beta1"
			},
			"tagline": "You Know, for Search"
		  }`,
	)
}

func testDiscover(t *testing.T, expectedVersion ESVersion, getRootResult string) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client := client.Mock(func(req *http.Request, resp *http.Response) error {
		resp.Body = io.NopCloser(strings.NewReader(getRootResult))
		return nil
	})

	version, err := Discover(ctx, client)
	assert.NoError(t, err)
	assert.Equal(t, expectedVersion, version)
}
