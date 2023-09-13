package pkgmanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

type packagistDistJSON struct {
	URL       string `json:"url"`
	Type      string `json:"type"`
	Shasum    string `json:"shasum,omitempty"`
	Reference string `json:"reference"`
}

func (d *packagistDistJSON) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "null":
		return nil
	case `"__unset"`:
		return nil
	}
	type raw packagistDistJSON
	return json.Unmarshal(data, (*raw)(d))
}

type packagistJSON struct {
	Packages map[string][]struct {
		Version           string            `json:"version"`
		VersionNormalized string            `json:"version_normalized"`
		License           []string          `json:"license,omitempty"`
		Time              time.Time         `json:"time"`
		Name              string            `json:"name,omitempty"`
		Dist              packagistDistJSON `json:"dist"`
	} `json:"packages"`
}

func getPackagistLatest(pkg string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://repo.packagist.org/p2/%s.json", pkg))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var details packagistJSON
	err = decoder.Decode(&details)
	if err != nil {
		return "", err
	}

	latestVersion := ""
	var lastTime time.Time
	for _, versions := range details.Packages {
		for _, v := range versions {
			if v.Time.Before(lastTime) {
				continue
			}
			lastTime = v.Time
			latestVersion = v.Version
		}
	}

	return latestVersion, nil
}

func getPackagistArchiveURL(pkgName, version string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://repo.packagist.org/p2/%s.json", pkgName))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var details packagistJSON
	err = decoder.Decode(&details)
	if err != nil {
		return "", err
	}

	for _, versions := range details.Packages {
		for _, v := range versions {
			if v.Version == version {
				return v.Dist.URL, nil
			}
		}
	}

	return "", nil
}

func getPackagistArchiveFilename(pkgName, version, _ string) string {
	pkg := strings.Split(pkgName, "/")
	return strings.Join([]string{pkg[0], "-", pkg[1], "-", version, ".zip"}, "")
}

var packagistPkgManager = PkgManager{
	ecosystem:       pkgecosystem.Packagist,
	latestVersion:   getPackagistLatest,
	archiveURL:      getPackagistArchiveURL,
	archiveFilename: getPackagistArchiveFilename,
}
