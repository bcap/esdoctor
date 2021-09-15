package version

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"esdoctor/client"
)

// ES uses semantic versioning https://semver.org/
type ESVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

func (v ESVersion) Set() bool {
	return v.Major != 0 || v.Minor != 0 || v.Patch != 0
}

func (v ESVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Major, v.Patch)
}

func (v ESVersion) MarshalJSON() ([]byte, error) {
	// The alias type is used here to copy all of the ESVersion data but
	// have no associated methods. This is necessary or else this MarshalJSON
	// method will be called in json.Marshal, causing an infinite recursion
	type Alias ESVersion

	withExtraFields := struct {
		Alias
		String string `json:"human"`
	}{
		Alias:  Alias(v),
		String: fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch),
	}
	return json.Marshal(withExtraFields)
}

func Discover(ctx context.Context, client client.Versioned) (ESVersion, error) {
	errResult := func(err error) (ESVersion, error) {
		return ESVersion{}, fmt.Errorf("failed to discover the ES version: %w", err)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		return errResult(err)
	}
	resp, err := client.Do(req)

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

	result, err := Parse(decoded.Version.Number)
	if err != nil {
		return errResult(err)
	}
	return result, nil
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
