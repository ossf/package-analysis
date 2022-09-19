package pkgecosystem

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type cratesIoJSON struct {
	Versions []struct {
		Num string `json:"num"`
	} `json:"versions"`
}

func getCratesIoLatest(pkg string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://crates.io/api/v1/crates/%s/versions", pkg))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var details cratesIoJSON
	err = decoder.Decode(&details)
	if err != nil {
		return "", err
	}

	return details.Versions[0].Num, nil
}

var cratesIoPkgManager = PkgManager{
	name:    "crates.io",
	image:   "gcr.io/ossf-malware-analysis/cratesio",
	command: "/usr/local/bin/analyze.py",
	latest:  getCratesIoLatest,
	dynamicPhases: []string{
		"install",
	},
}
