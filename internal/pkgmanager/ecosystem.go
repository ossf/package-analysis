package pkgmanager

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

var ErrNoArchiveURL = errors.New("archive URL not found")

// PkgManager represents how packages from a common ecosystem are accessed.
type PkgManager struct {
	ecosystem       pkgecosystem.Ecosystem
	latestVersion   func(name string) (string, error)
	archiveURL      func(name, version string) (string, error)
	archiveFilename func(name, version, downloadURL string) string
	extractArchive  func(path, outputDir string) error
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

/*
defaultArchiveFilename returns a naive default choice of filename from a
download URL by simply returning everything after the final slash in the URL.
There is no guarantee that this method results in a good archive filename - e.g.
NPM package namespace is not always included in the URL-derived filename.
*/
func defaultArchiveFilename(_, _, downloadURL string) string {
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

/*
DownloadArchive downloads an archive of the given package name and version
to the specified directory, and returns the path to the downloaded archive.
The archive is named according to ecosystem-specific rules.

directory specifies the destination directory for the archive.
If an empty string is passed, the current directory is used.

If an error occurs during download of the file, it is returned along with
an empty path value.
*/
func (p *PkgManager) DownloadArchive(name, version, directory string) (string, error) {
	if directory == "" {
		directory = "."
	}

	downloadURL, err := p.archiveURL(name, version)
	if err != nil {
		return "", err
	}
	if downloadURL == "" {
		return "", fmt.Errorf("%w: package %s @ %s", ErrNoArchiveURL, name, version)
	}

	baseFilename := p.archiveFilename(name, version, downloadURL)
	if baseFilename == "" {
		panic("base filename for archive is empty")
	}

	destPath := filepath.Join(directory, baseFilename)
	if err := downloadToPath(destPath, downloadURL); err != nil {
		return "", err
	}

	return destPath, nil
}

func (p *PkgManager) ExtractArchive(archivePath, outputDir string) error {
	if p.extractArchive != nil {
		return p.extractArchive(archivePath, outputDir)
	}
	return fmt.Errorf("archive extraction not implemented for %s", p.Ecosystem())
}

func normalizePkgName(pkg string) string {
	return strings.ToLower(pkg)
}
