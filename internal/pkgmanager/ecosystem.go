package pkgmanager

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/ossf/package-analysis/internal/utils"
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

/*
DownloadArchive downloads an archive of the given package name and version
to the specified directory, and returns the path to the downloaded archive.

If directory is empty, the current directory is used.

If renameWithHash is true, the SHA256 sum of the archive is computed via
utils.HashFile and appended to the base filename, separated by '-'.

Note: If an error occurs during hashing or renaming of the file, it is not removed.
The temporary path is returned; it is the responsibility of the caller
to remove the file and retry if desired.

If an error occurs during download of the file, the error is returned along
with an empty path value. No cleanup is required.
*/
func (p *PkgManager) DownloadArchive(name, version, directory string, renameWithHash bool) (string, error) {
	if directory == "" {
		directory = "."
	}

	downloadURL, err := p.archiveURL(name, version)
	if err != nil {
		return "", err
	}
	if downloadURL == "" {
		return "", fmt.Errorf("no url found for package %s, version %s", name, version)
	}

	baseFilename := p.archiveFilename(name, version, downloadURL)
	if baseFilename == "" {
		panic("base filename for archive is empty")
	}

	// final path if renameWithHash is false
	destPath := filepath.Join(directory, baseFilename)
	if err := downloadToPath(destPath, downloadURL); err != nil {
		return "", err
	}

	if !renameWithHash {
		// nothing more to do
		return destPath, nil
	}

	// TODO this puts the hash after the file extension - see if we can put it before
	pathWithHash, err := utils.RenameWithHash(destPath, baseFilename+"-", "")
	if err != nil {
		// rename failed but don't remove the file, just return unmodified path
		return destPath, err
	}

	return pathWithHash, nil
}

func (p *PkgManager) ExtractArchive(archivePath, outputDir string) error {
	return p.extractArchive(archivePath, outputDir)
}

func normalizePkgName(pkg string) string {
	return strings.ToLower(pkg)
}
