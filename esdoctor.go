package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	es8 "github.com/elastic/go-elasticsearch/v8"
)

func main() {
	err := Diagnose(
		context.Background(),
		"https://vpc-ds-10-logistics3-6vmimxrf5vlhagzjntvrrrlxyq.us-east-1.es.amazonaws.com/",
	)
	if err != nil {
		log.Errorf("Execution failed: %v", err)
		os.Exit(1)
	}
}

type options struct {
}

func newOptions(optionFns ...Option) options {
	options := options{}
	for _, fn := range optionFns {
		fn(&options)
	}
	return options
}

type Option func(*options)

func Diagnose(ctx context.Context, endpoint string, optionFns ...Option) error {
	options := newOptions(optionFns...)
	log.Debugf("Running diagnostics on endpoint %s with the following options: %+v", options)

	version, err := FetchESVersion(ctx, endpoint)
	if err != nil {
		return err
	}
	fmt.Println(version)
	return nil
}

// ES uses semantic versioning https://semver.org/
type ESVersion struct {
	Major int
	Minor int
	Patch int
}

func FetchESVersion(ctx context.Context, endpoint string) (ESVersion, error) {
	client, err := es7.NewClient(es8.Config{Addresses: []string{endpoint}})
	if err != nil {
		return ESVersion{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return ESVersion{}, err
	}
	resp, err := client.Perform(req)
	if err != nil {
		return ESVersion{}, err
	}
	if resp.StatusCode != 200 {
		return ESVersion{}, fmt.Errorf("got http status code %d while doing GET /", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ESVersion{}, err
	}

	var parsed struct {
		Version struct {
			Number string `json:"number"`
		} `json:"version"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return ESVersion{}, fmt.Errorf("failed to json parse the response from GET /")
	}

	version := parsed.Version.Number
	split := strings.Split(version, ".")
	if len(split) != 3 {
		return ESVersion{}, fmt.Errorf(
			"invalid version returned from ES: %s. Expected 3 fields sepparated by dot (.)",
			version,
		)
	}

	major, err := strconv.Atoi(split[0])
	if err != nil {
		return ESVersion{}, fmt.Errorf("ES major version %q from %q is not a number", split[0], version)
	}
	minor, err := strconv.Atoi(split[1])
	if err != nil {
		return ESVersion{}, fmt.Errorf("ES minor version %q from %q is not a number", split[1], version)
	}
	patch, err := strconv.Atoi(split[2])
	if err != nil {
		return ESVersion{}, fmt.Errorf("ES patch version %q from %q is not a number", split[2], version)
	}

	return ESVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}
