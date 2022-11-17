package pkgecosystem

import (
	"fmt"
	"strings"
)

// Ecosystem represents an open source package ecosystem from which packages can be downloaded
type Ecosystem string

// RunPhase represents a way to 'run' a package during its usage lifecycle.
// This is relevant to dynamic analysis.
type RunPhase string

const (
	Import  RunPhase = "import"
	Install RunPhase = "install"

	CratesIO  Ecosystem = "crates.io"
	NPM       Ecosystem = "npm"
	Packagist Ecosystem = "packagist"
	PyPi      Ecosystem = "pypi"
	Rubygems  Ecosystem = "rubygems"
)

// PkgManager
// Represents how packages from a common ecosystem are accessed
type PkgManager struct {
	ecosystem  Ecosystem
	image      string
	command    string
	latest     func(string) (string, error)
	archiveUrl func(string, string) (string, error)
	runPhases  []RunPhase
}

var (
	supportedPkgManagers = map[Ecosystem]*PkgManager{
		npmPkgManager.ecosystem:       &npmPkgManager,
		pypiPkgManager.ecosystem:      &pypiPkgManager,
		rubygemsPkgManager.ecosystem:  &rubygemsPkgManager,
		packagistPkgManager.ecosystem: &packagistPkgManager,
		cratesPkgManager.ecosystem:    &cratesPkgManager,
	}
)

func Manager(ecosystem string) *PkgManager {
	return supportedPkgManagers[Ecosystem(ecosystem)]
}

// String implements the Stringer interface to support pretty printing.
func (p *PkgManager) String() string {
	return string(p.ecosystem)
}

func (p *PkgManager) DynamicAnalysisImage() string {
	return p.image
}

func (p *PkgManager) RunPhases() []RunPhase {
	return p.runPhases
}

func (p *PkgManager) Latest(name string) (*Pkg, error) {
	name = normalizePkgName(name)
	version, err := p.latest(name)
	if err != nil {
		return nil, err
	}
	return &Pkg{
		name:    name,
		version: version,
		manager: p,
	}, nil
}

func (p *PkgManager) Local(name, version, localPath string) *Pkg {
	return &Pkg{
		name:    normalizePkgName(name),
		version: version,
		local:   localPath,
		manager: p,
	}
}

func (p *PkgManager) Package(name, version string) *Pkg {
	return &Pkg{
		name:    normalizePkgName(name),
		version: version,
		manager: p,
	}
}

func (p *PkgManager) DownloadArchive(name, version, directory string) (string, error) {
	if directory == "" {
		return "", fmt.Errorf("no directory specified")
	}

	downloadURL, err := p.archiveUrl(name, version)
	if err != nil {
		return "", err
	}

	archivePath, err := downloadToDirectory(directory, downloadURL)
	if err != nil {
		return "", err
	}

	return archivePath, nil
}

func normalizePkgName(pkg string) string {
	return strings.ToLower(pkg)
}
