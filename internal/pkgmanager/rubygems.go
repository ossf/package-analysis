package pkgmanager

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
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

func getRubyGemsArchiveURL(pkgName, version string) (string, error) {
	pkgURL := fmt.Sprintf("https://rubygems.org/gems/%v-%v.gem", pkgName, version)
	resp, err := http.Get(pkgURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return pkgURL, nil
}

var rubygemsPkgManager = PkgManager{
	ecosystem:       pkgecosystem.RubyGems,
	latestVersion:   getRubyGemsLatest,
	archiveURL:      getRubyGemsArchiveURL,
	archiveFilename: defaultArchiveFilename,
}
