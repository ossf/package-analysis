package pkgecosystem

import (
	"fmt"
	"strings"
)

// RunPhase represents a way to 'run' a package during its usage lifecycle
// This is relevant to dynamic analysis
type RunPhase string

// Ecosystem denotes a package ecosystem
type Ecosystem string

const (
	Import  RunPhase = "import"
	Install RunPhase = "install"

	CratesIO  Ecosystem = "crates.io"
	NPM       Ecosystem = "npm"
	Packagist Ecosystem = "packagist"
	PyPi      Ecosystem = "pypi"
	Rubygems  Ecosystem = "rubygems"
)

func IsSupportedEcosystem(ecosystemName string) bool {
	return getPackageFactory(Ecosystem(ecosystemName)) != nil
}

// packageFactory groups together ecosystem-specific logic for querying package metadata from remote repositories
// and local files, and handles creation of Package instances for its associated ecosystem
type packageFactory interface {
	// Ecosystem returns the name of the package ecosystem associated with this packageFactory, i.e. the type of
	// package that it will create
	Ecosystem() Ecosystem

	// constructPackage constructs a concrete Package object corresponding to this ecosystem; no checks are done
	constructPackage(name, version, localPath string) Package
	// getLatestVersion retrieves the latest available version for the given package name
	getLatestVersion(name string) (string, error)
}

func getPackageFactory(ecosystem Ecosystem) packageFactory {
	switch ecosystem {
	case CratesIO:
		return cratesPkgFactory{}
	case NPM:
		return npmPkgFactory{}
	case Packagist:
		return packagistPkgFactory{}
	case PyPi:
		return pypiPkgFactory{}
	case Rubygems:
		return rubygemsPkgFactory{}
	default:
		return nil
	}
}

// MakePackage initialises a package record
// If version == "", the latest available version for the package will be fetched and used.
// If localPath != "", the file pointed to by that path is assumed to be an archive containing the package.
func MakePackage(ecosystem, name, version, localPath string) (pkg Package, err error) {
	factory := getPackageFactory(Ecosystem(ecosystem))
	if factory == nil {
		return nil, fmt.Errorf("unsupported package ecosystem: " + string(ecosystem))
	}

	name = normalizePkgName(name)

	needLatestVersion := localPath == "" && version == ""
	if needLatestVersion {
		version, err = factory.getLatestVersion(name)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest version for package %s: %v", name, err)
		}
	} else {
		// TODO check if desired package name/version exists in ecosystem, before creating package.
		//  If a local package file is specified, the version argument given here should be
		//  cross-checked with actual version parsed from package file

		//err = pkg.Validate()
		err = nil
	}

	pkg = factory.constructPackage(name, version, localPath)

	return
}

func normalizePkgName(pkg string) string {
	return strings.ToLower(pkg)
}
