package pkgmanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

func getCratesArchiveURL(pkgName, version string) (string, error) {
	pkgURL := fmt.Sprintf("https://crates.io/api/v1/crates/%s/%s/download", pkgName, version)
	resp, err := http.Get(pkgURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return pkgURL, nil
}

func getCratesArchiveFilename(pkgName, version, _ string) string {
	return strings.Join([]string{pkgName, "-", version, ".tar.gz"}, "")
}

var cratesPkgManager = PkgManager{
	ecosystem:       pkgecosystem.CratesIO,
	latestVersion:   getCratesLatest,
	archiveURL:      getCratesArchiveURL,
	archiveFilename: getCratesArchiveFilename,
}
