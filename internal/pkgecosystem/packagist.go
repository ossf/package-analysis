package pkgecosystem

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type packagistJSON struct {
	Packages map[string][]struct {
		Version           string    `json:"version"`
		VersionNormalized string    `json:"version_normalized"`
		License           []string  `json:"license,omitempty"`
		Time              time.Time `json:"time"`
		Name              string    `json:"name,omitempty"`
	} `json:"packages"`
}

func getPackagistLatest(pkg string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://repo.packagist.org/p2/%s.json", pkg))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var details packagistJSON
	err = decoder.Decode(&details)
	if err != nil {
		return "", err
	}

	latestVersion := ""
	var lastTime time.Time
	for _, versions := range details.Packages {
		for _, v := range versions {
			if v.Time.Before(lastTime) {
				continue
			}
			lastTime = v.Time
			latestVersion = v.Version
		}
	}

	return latestVersion, nil
}

type packagistPkgFactory struct{}

func (f packagistPkgFactory) Ecosystem() Ecosystem {
	return Packagist
}

func (f packagistPkgFactory) constructPackage(name, version, localPath string) Package {
	return packagistPkg{name, version, localPath}
}

func (f packagistPkgFactory) getLatestVersion(name string) (string, error) {
	return getPackagistLatest(name)
}

type packagistPkg struct {
	name      string
	version   string
	localPath string
}

func (p packagistPkg) Name() string {
	return p.name
}

func (p packagistPkg) Ecosystem() string {
	return string(Packagist)
}

func (p packagistPkg) Version() string {
	return p.version
}

func (p packagistPkg) LocalPath() string {
	return p.localPath
}

func (p packagistPkg) Download() (string, error) {
	notImplemented()
	return "", nil
}

func (p packagistPkg) Command(phase RunPhase) []string {
	return phaseCommand(p, phase)
}

func (p packagistPkg) DynamicAnalysisImage() string {
	return "gcr.io/ossf-malware-analysis/packagist"
}

func (p packagistPkg) DynamicRunPhases() []RunPhase {
	return []RunPhase{Install, Import}
}

func (p packagistPkg) String() string {
	return packageToString(p)
}

func (p packagistPkg) Validate() error {
	notImplemented()
	return nil
}

func (p packagistPkg) baseCommand() string {
	return "/usr/local/bin/analyse.php"
}
