package pkgmanager

import (
	"fmt"
	"path"
	"strings"

	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

// PkgManager represents how packages from a common ecosystem are accessed.
type PkgManager struct {
	ecosystem       pkgecosystem.Ecosystem
	latestVersion   func(string) (string, error)
	archiveURL      func(string, string) (string, error)
	archiveFilename func(string, string, string) string
	extractArchive  func(string, string) error
}

var (
	supportedPkgManagers = map[pkgecosystem.Ecosystem]*PkgManager{
		npmPkgManager.ecosystem:       &npmPkgManager,
		pypiPkgManager.ecosystem:      &pypiPkgManager,
		rubygemsPkgManager.ecosystem:  &rubygemsPkgManager,
		packagistPkgManager.ecosystem: &packagistPkgManager,
		cratesPkgManager.ecosystem:    &cratesPkgManager,
	}
)

func DefaultArchiveFilename(pkgName, version, downloadURL string) string {
	return path.Base(downloadURL)
}

func Manager(e pkgecosystem.Ecosystem) *PkgManager {
	return supportedPkgManagers[e]
}

// String implements the Stringer interface to support pretty printing.
func (p *PkgManager) String() string {
	return string(p.ecosystem)
}

func (p *PkgManager) Ecosystem() pkgecosystem.Ecosystem {
	return p.ecosystem
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
	if downloadURL == "" {
		return "", fmt.Errorf("no url found for package %s, version %s", name, version)
	}

	fileName := p.archiveFilename(name, version, downloadURL)

	archivePath, err := downloadToDirectory(directory, downloadURL, fileName)

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
