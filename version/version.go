package version

import (
	"context"
	"encoding/json"
	"esdoctor/client"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// ES uses semantic versioning https://semver.org/
type ESVersion struct {
	Major int
	Minor int
	Patch int
}

func Discover(ctx context.Context, client client.Versioned) (ESVersion, error) {
	errResult := func(err error) (ESVersion, error) {
		return ESVersion{}, fmt.Errorf("failed to discover the ES version: %w", err)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		return errResult(err)
	}
	resp, err := client.V8.Transport.Perform(req)

	if err != nil {
		return errResult(err)
	}

	var decoded struct {
		Version struct {
			Number string `json:"number"`
		} `json:"version"`
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errResult(err)
	}
	if err := json.Unmarshal(bodyBytes, &decoded); err != nil {
		return errResult(err)
	}

	return Parse(decoded.Version.Number)
}

func Parse(version string) (ESVersion, error) {
	errResult := func() (ESVersion, error) {
		return ESVersion{}, fmt.Errorf("invalid version %q: must contain 3 numbers joined by . (dot)", version)
	}

	split := strings.Split(version, ".")
	if len(split) != 3 {
		return errResult()
	}

	major, errMajor := strconv.Atoi(split[0])
	minor, errMinor := strconv.Atoi(split[1])
	patch, errPatch := strconv.Atoi(split[2])
	if errMajor != nil || errMinor != nil || errPatch != nil {
		return errResult()
	}

	return ESVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}
