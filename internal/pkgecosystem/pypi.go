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

type pypiPkgFactory struct{}

func (f pypiPkgFactory) Ecosystem() Ecosystem {
	return PyPi
}

func (f pypiPkgFactory) constructPackage(name, version, localPath string) Package {
	return pypiPkg{name, version, localPath}
}

func (f pypiPkgFactory) getLatestVersion(name string) (string, error) {
	return getPyPILatest(name)
}

type pypiPkg struct {
	name      string
	version   string
	localPath string
}

func (p pypiPkg) Name() string {
	return p.name
}

func (p pypiPkg) Ecosystem() string {
	return string(PyPi)
}

func (p pypiPkg) Version() string {
	return p.version
}

func (p pypiPkg) LocalPath() string {
	return p.localPath
}

func (p pypiPkg) Download() (string, error) {
	notImplemented()
	return "", nil
}

func (p pypiPkg) Command(phase RunPhase) []string {
	return phaseCommand(p, phase)
}

func (p pypiPkg) DynamicAnalysisImage() string {
	return "gcr.io/ossf-malware-analysis/python"
}

func (p pypiPkg) DynamicRunPhases() []RunPhase {
	return []RunPhase{Install, Import}
}

func (p pypiPkg) String() string {
	return packageToString(p)
}

func (p pypiPkg) Validate() error {
	notImplemented()
	return nil
}

func (p pypiPkg) baseCommand() string {
	return "/usr/local/bin/analyse.py"
}
