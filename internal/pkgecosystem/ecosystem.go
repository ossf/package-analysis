package pkgecosystem

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/pkg/api"
)

// PkgManager
// Represents how packages from a common ecosystem are accessed
type PkgManager struct {
	ecosystem      api.Ecosystem
	image          string
	command        string
	unifiedCommand string
	latestVersion  func(string) (string, error)
	archiveUrl     func(string, string) (string, error)
	extractArchive func(string, string) error
	runPhases      []api.RunPhase
}

var (
	supportedPkgManagers = map[api.Ecosystem]*PkgManager{
		npmPkgManager.ecosystem:       &npmPkgManager,
		pypiPkgManager.ecosystem:      &pypiPkgManager,
		rubygemsPkgManager.ecosystem:  &rubygemsPkgManager,
		packagistPkgManager.ecosystem: &packagistPkgManager,
		cratesPkgManager.ecosystem:    &cratesPkgManager,
	}
)

func Manager(ecosystem api.Ecosystem) *PkgManager {
	return supportedPkgManagers[ecosystem]
}

// String implements the Stringer interface to support pretty printing.
func (p *PkgManager) String() string {
	return string(p.ecosystem)
}

func (p *PkgManager) DynamicAnalysisImage() string {
	return p.image
}

func (p *PkgManager) RunPhases() []api.RunPhase {
	return p.runPhases
}

func (p *PkgManager) Latest(name string) (*Pkg, error) {
	name = normalizePkgName(name)
	version, err := p.latestVersion(name)
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

func (p *PkgManager) ExtractArchive(archivePath, outputDir string) error {
	return p.extractArchive(archivePath, outputDir)
}

func normalizePkgName(pkg string) string {
	return strings.ToLower(pkg)
}
