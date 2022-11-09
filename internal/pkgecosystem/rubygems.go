package pkgecosystem

import (
	"encoding/json"
	"fmt"
	"net/http"
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

type rubygemsPkgFactory struct{}

func (f rubygemsPkgFactory) Ecosystem() Ecosystem {
	return Rubygems
}

func (f rubygemsPkgFactory) constructPackage(name, version, localPath string) Package {
	return rubygemsPkg{name, version, localPath}
}

func (f rubygemsPkgFactory) getLatestVersion(name string) (string, error) {
	return getRubyGemsLatest(name)
}

type rubygemsPkg struct {
	name      string
	version   string
	localPath string
}

func (p rubygemsPkg) Name() string {
	return p.name
}

func (p rubygemsPkg) Ecosystem() string {
	return string(Rubygems)
}

func (p rubygemsPkg) Version() string {
	return p.version
}

func (p rubygemsPkg) LocalPath() string {
	return p.localPath
}

func (p rubygemsPkg) Download() (string, error) {
	notImplemented()
	return "", nil
}

func (p rubygemsPkg) Command(phase RunPhase) []string {
	return phaseCommand(p, phase)
}

func (p rubygemsPkg) DynamicAnalysisImage() string {
	return "gcr.io/ossf-malware-analysis/ruby"
}

func (p rubygemsPkg) DynamicRunPhases() []RunPhase {
	return []RunPhase{Install, Import}
}

func (p rubygemsPkg) String() string {
	return packageToString(p)
}

func (p rubygemsPkg) Validate() error {
	notImplemented()
	return nil
}

func (p rubygemsPkg) baseCommand() string {
	return "/usr/local/bin/analyse.rb"
}
