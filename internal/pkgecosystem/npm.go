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

type npmPkgFactory struct{}

func (f npmPkgFactory) Ecosystem() Ecosystem {
	return NPM
}

func (f npmPkgFactory) constructPackage(name, version, localPath string) Package {
	return npmPkg{name, version, localPath}
}

func (f npmPkgFactory) getLatestVersion(name string) (string, error) {
	return getNPMLatest(name)
}

type npmPkg struct {
	name      string
	version   string
	localPath string
}

func (p npmPkg) Name() string {
	return p.name
}

func (p npmPkg) Ecosystem() string {
	return string(NPM)
}

func (p npmPkg) Version() string {
	return p.version
}

func (p npmPkg) LocalPath() string {
	return p.localPath
}

func (p npmPkg) Download() (string, error) {
	notImplemented()
	return "", nil
}

func (p npmPkg) Command(phase RunPhase) []string {
	return phaseCommand(p, phase)
}

func (p npmPkg) DynamicAnalysisImage() string {
	return "gcr.io/ossf-malware-analysis/node"
}

func (p npmPkg) DynamicRunPhases() []RunPhase {
	return []RunPhase{Install, Import}
}

func (p npmPkg) String() string {
	return packageToString(p)
}

func (p npmPkg) baseCommand() string {
	return "/usr/local/bin/analyse.js"
}
