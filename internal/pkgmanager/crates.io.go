package pkgmanager

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ossf/package-analysis/pkg/api/analysisrun"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

type cratesJSON struct {
	Versions []struct {
		Num string `json:"num"`
	} `json:"versions"`
}

func getCratesLatest(pkg string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://crates.io/api/v1/crates/%s/versions", pkg))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var details cratesJSON
	err = decoder.Decode(&details)
	if err != nil {
		return "", err
	}

	return details.Versions[0].Num, nil
}

var cratesPkgManager = PkgManager{
	ecosystem:     pkgecosystem.CratesIO,
	image:         "gcr.io/ossf-malware-analysis/crates.io",
	command:       "/usr/local/bin/analyze.py",
	latestVersion: getCratesLatest,
	dynamicPhases: analysisrun.DefaultDynamicPhases(),
}

var cratesPkgManagerCombinedSandbox = PkgManager{
	ecosystem:     pkgecosystem.CratesIO,
	image:         combinedDynamicAnalysisImage,
	command:       "/usr/local/bin/analyze-rust.py",
	latestVersion: getCratesLatest,
	dynamicPhases: analysisrun.DefaultDynamicPhases(),
}
