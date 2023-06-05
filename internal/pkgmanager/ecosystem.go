package pkgmanager

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/pkg/api/analysisrun"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

// PkgManager represents how packages from a common ecosystem are accessed.
type PkgManager struct {
	ecosystem      pkgecosystem.Ecosystem
	image          string
	command        string
	latestVersion  func(string) (string, error)
	archiveURL     func(string, string) (string, error)
	extractArchive func(string, string) error
	dynamicPhases  []analysisrun.DynamicPhase
}

const combinedDynamicAnalysisImage = "gcr.io/ossf-malware-analysis/dynamic-analysis"

var (
	supportedPkgManagers = map[pkgecosystem.Ecosystem]*PkgManager{
		npmPkgManager.ecosystem:       &npmPkgManager,
		pypiPkgManager.ecosystem:      &pypiPkgManager,
		rubygemsPkgManager.ecosystem:  &rubygemsPkgManager,
		packagistPkgManager.ecosystem: &packagistPkgManager,
		cratesPkgManager.ecosystem:    &cratesPkgManager,
	}
)

func Manager(e pkgecosystem.Ecosystem) *PkgManager {
	return supportedPkgManagers[e]
}

// String implements the Stringer interface to support pretty printing.
func (p *PkgManager) String() string {
	return string(p.ecosystem)
}

func (p *PkgManager) DynamicAnalysisImage() string {
	return p.image
}

func (p *PkgManager) DynamicPhases() []analysisrun.DynamicPhase {
	return p.dynamicPhases
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

	downloadURL, err := p.archiveURL(name, version)
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
