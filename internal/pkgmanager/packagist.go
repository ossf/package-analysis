package pkgmanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ossf/package-analysis/pkg/api/analysisrun"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

type packagistJSON struct {
	Packages map[string][]struct {
		Version           string    `json:"version"`
		VersionNormalized string    `json:"version_normalized"`
		License           []string  `json:"license,omitempty"`
		Time              time.Time `json:"time"`
		Name              string    `json:"name,omitempty"`
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

var packagistPkgManager = PkgManager{
	ecosystem:     pkgecosystem.Packagist,
	image:         "gcr.io/ossf-malware-analysis/packagist",
	command:       "/usr/local/bin/analyze.php",
	latestVersion: getPackagistLatest,
	dynamicPhases: analysisrun.DefaultDynamicPhases(),
}

var packagistPkgManagerCombinedSandbox = PkgManager{
	ecosystem:     pkgecosystem.Packagist,
	image:         combinedDynamicAnalysisImage,
	command:       "/usr/local/bin/analyze-php.php",
	latestVersion: getPackagistLatest,
	dynamicPhases: analysisrun.DefaultDynamicPhases(),
}
