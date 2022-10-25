package pkgecosystem

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type pypiJSON struct {
	Info struct {
		Version string `json:"version"`
	} `json:"info"`
}

func getPyPILatest(pkg string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://pypi.org/pypi/%s/json", pkg))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var details pypiJSON
	err = decoder.Decode(&details)
	if err != nil {
		return "", err
	}

	return details.Info.Version, nil
}

var pypiPkgManager = PkgManager{
	name:    "pypi",
	image:   "gcr.io/ossf-malware-analysis/python",
	command: "/usr/local/bin/analyze.py",
	latest:  getPyPILatest,
	runPhases: []RunPhase{
		Install,
		Import,
	},
}
