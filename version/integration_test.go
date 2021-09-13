// +build integration

// to run tests in this package, go test needs to receive the integration tag. Eg:
//   go test -tags=integration

package version

import (
	"context"
	"testing"
	"time"

	"esdoctor/client"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationDiscover(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client, err := client.New("http://localhost:9200")
	assert.NoError(t, err)

	version, err := Discover(ctx, client)
	assert.NoError(t, err)
	assert.NotEqual(t, ESVersion{}, version)
}
