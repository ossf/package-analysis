package pkgmanager

import (
	"fmt"
	"path"
	"strings"

	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

// PkgManager represents how packages from a common ecosystem are accessed.
type PkgManager struct {
	ecosystem      pkgecosystem.Ecosystem
	latestVersion  func(string) (string, error)
	archiveURL     func(string, string) (string, error)
	extractArchive func(string, string) error
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

func (p *PkgManager) DownloadArchive(name, version, directory string, append_hash bool) (string, error) {
	if directory == "" {
		return "", fmt.Errorf("no directory specified")
	}

	if p.archiveURL == nil {
		return "", fmt.Errorf("not yet implemented for %s", p.Ecosystem())
	}

	downloadURL, err := p.archiveURL(name, version)
	if err != nil {
		return "", err
	}

	fileName := ""
	// Rename for crates (The name of downloaded archive is download)
	if p.ecosystem == "crates.io" && path.Base(downloadURL) == "download" {
		fileName = strings.Join([]string{name, "-", version, ".tar.gz"}, "")
	} else if p.ecosystem == "packagist" {
		pkg := strings.Split(name, "/")
		fileName = strings.Join([]string{pkg[0], "-", pkg[1], "-", version, ".zip"}, "")
	}

	archivePath, err := downloadToDirectory(directory, downloadURL, fileName, append_hash)

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
