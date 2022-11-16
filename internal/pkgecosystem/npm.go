package pkgecosystem

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// npmPackageJSON represents relevant JSON data from the NPM registry response
// when package information is requested.
// See https://github.com/npm/registry/blob/master/docs/responses/package-metadata.md
type npmPackageJSON struct {
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
}

// npmVersionJSON represents relevant JSON data from the NPM registry response
// when package version information is requested.
// See https://github.com/npm/registry/blob/master/docs/responses/package-metadata.md
type npmVersionJSON struct {
	Dist struct {
		Tarball string `json:"tarball"`
	} `json:"dist"`
}

func getNPMLatest(pkg string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", pkg))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var details npmPackageJSON
	err = decoder.Decode(&details)
	if err != nil {
		return "", err
	}

	return details.DistTags.Latest, nil
}

func downloadNPMArchive(pkgName string, version string, directory string) (path string, err error) {
	funcName := "downloadNPMArchive"
	if directory == "" {
		return "", fmt.Errorf("%s: no directory specified", funcName)
	}

	url := fmt.Sprintf("https://registry.npmjs.org/%s/%s", pkgName, version)

	resp, err := http.Get(url)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("%s: error reading HTTP response: %v", funcName, err)
	}

	responseString := string(responseBytes)
	decoder := json.NewDecoder(strings.NewReader(responseString))
	var details npmVersionJSON
	err = decoder.Decode(&details)
	if err != nil {
		return "", fmt.Errorf("%s: %v. Response: %s", funcName, err, responseString)
	}

	downloadURL := details.Dist.Tarball

	// get filename from last element of NPM download url
	archivePath, err := downloadToDirectory(directory, downloadURL)

	if err != nil {
		return "", err
	}

	return archivePath, nil
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
