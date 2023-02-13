package pkgecosystem

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/pkg/api"
)

// pypiPackageInfoJSON represents relevant JSON data from the PyPI web API response
// when package information is requested. The differences in response format between
// (valid) requests made with a specific package version and with no package version
// are not significant in our case.
// (In particular, if the request contains a valid version, Urls contains a single entry
// holding information for that package version. If the version is unspecified, Urls contains
// an entry corresponding to each version of the package available on PyPI.)
// See https://warehouse.pypa.io/api-reference/json.html and https://peps.python.org/pep-0691
type pypiPackageInfoJSON struct {
	Info struct {
		Version string `json:"version"`
	} `json:"info"`
	Urls []struct {
		PackageType string `json:"packagetype"`
		Url         string `json:"url"`
	} `json:"urls"`
}

func getPyPILatest(pkg string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://pypi.org/pypi/%s/json", pkg))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var details pypiPackageInfoJSON
	err = decoder.Decode(&details)
	if err != nil {
		return "", err
	}

	return details.Info.Version, nil
}

func getPyPIArchiveURL(pkgName, version string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://pypi.org/pypi/%s/%s/json", pkgName, version))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading HTTP response: %v", err)
	}

	responseString := string(responseBytes)
	decoder := json.NewDecoder(strings.NewReader(responseString))
	var packageInfo pypiPackageInfoJSON
	err = decoder.Decode(&packageInfo)
	if err != nil {
		// invalid version, non-existent package, etc. Details in responseString
		return "", fmt.Errorf("%v. PyPI response: %s", err, responseString)
	}

	// Need to find the archive with PackageType == "sdist"
	for _, url := range packageInfo.Urls {
		if url.PackageType == "sdist" {
			return url.Url, nil
		}
	}
	// can't find source tarball
	return "", fmt.Errorf("source tarball not found for %s, version %s", pkgName, version)

}

var pypiPkgManager = PkgManager{
	ecosystem:      api.EcosystemPyPI,
	image:          "gcr.io/ossf-malware-analysis/python",
	command:        "/usr/local/bin/analyze.py",
	unifiedCommand: "/usr/local/bin/analyze-python.py",
	latestVersion:  getPyPILatest,
	archiveUrl:     getPyPIArchiveURL,
	extractArchive: utils.ExtractTarGzFile,
	runPhases: []api.RunPhase{
		api.RunPhaseInstall,
		api.RunPhaseImport,
	},
}
