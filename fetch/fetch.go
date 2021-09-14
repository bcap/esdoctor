package fetch

import (
	"context"
	"encoding/json"
	"fmt"

	"esdoctor/client"
)

func Fetch(ctx context.Context, client client.Versioned, api string, obj interface{}) error {
	resp, err := client.Request(ctx, "GET", api, nil, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to fetch %s, got status code %d from ES", api, resp.StatusCode)
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&obj); err != nil {
		return fmt.Errorf("failed to json decode the response from %s: %w", api, err)
	}
	return nil
}
