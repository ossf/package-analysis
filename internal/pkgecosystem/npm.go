package pkgecosystem

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type npmJSON struct {
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
}

func getNPMLatest(pkg string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", pkg))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var details npmJSON
	err = decoder.Decode(&details)
	if err != nil {
		return "", err
	}

	return details.DistTags.Latest, nil
}

var npmPkgManager = PkgManager{
	name:    "npm",
	image:   "gcr.io/ossf-malware-analysis/node",
	command: "/usr/local/bin/analyze.js",
	latest:  getNPMLatest,
	runPhases: []RunPhase{
		Install,
		Import,
	},
}
