package pkgecosystem

import (
	"encoding/json"
	"fmt"
	"net/http"
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

type cratesPkgFactory struct{}

func (f cratesPkgFactory) Ecosystem() Ecosystem {
	return CratesIO
}

func (f cratesPkgFactory) constructPackage(name, version, localPath string) Package {
	return cratesPkg{name, version, localPath}
}

func (f cratesPkgFactory) getLatestVersion(name string) (string, error) {
	return getCratesLatest(name)
}

type cratesPkg struct {
	name      string
	version   string
	localPath string
}

func (p cratesPkg) Name() string {
	return p.name
}

func (p cratesPkg) Ecosystem() string {
	return string(NPM)
}

func (p cratesPkg) Version() string {
	return p.version
}

func (p cratesPkg) LocalPath() string {
	return p.localPath
}

func (p cratesPkg) Download() (string, error) {
	notImplemented()
	return "", nil
}

func (p cratesPkg) Command(phase RunPhase) []string {
	return phaseCommand(p, phase)
}

func (p cratesPkg) DynamicAnalysisImage() string {
	return "gcr.io/ossf-malware-analysis/crates.io"
}

func (p cratesPkg) DynamicRunPhases() []RunPhase {
	return []RunPhase{Install}
}

func (p cratesPkg) String() string {
	return packageToString(p)
}

func (p cratesPkg) Validate() error {
	notImplemented()
	return nil
}

func (p cratesPkg) baseCommand() string {
	return "/usr/local/bin/analyze.py"
}
