package pkgecosystem

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ossf/package-analysis/pkg/api"
)

type rubygemsJSON struct {
	Version string `json:"version"`
}

func getRubyGemsLatest(pkg string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://rubygems.org/api/v1/gems/%s.json", pkg))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var details rubygemsJSON
	err = decoder.Decode(&details)
	if err != nil {
		return "", err
	}

	return details.Version, nil
}

var rubygemsPkgManager = PkgManager{
	ecosystem:     api.EcosystemRubyGems,
	image:         "gcr.io/ossf-malware-analysis/ruby",
	command:       "/usr/local/bin/analyze.rb",
	latestVersion: getRubyGemsLatest,
	runPhases: []api.RunPhase{
		api.RunPhaseInstall,
		api.RunPhaseImport,
	},
}
