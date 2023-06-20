package worker

import (
	"fmt"

	"github.com/package-url/packageurl-go"

	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

// ResolvePkg creates a Pkg object with the arguments passed to the worker process.
func ResolvePkg(manager *pkgmanager.PkgManager, name, version, localPath string) (pkg *pkgmanager.Pkg, err error) {
	switch {
	case localPath != "":
		pkg = manager.Local(name, version, localPath)
	case version != "":
		pkg = manager.Package(name, version)
	default:
		pkg, err = manager.Latest(name)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest version: %w", err)
		}
		if pkg.Version() == "" {
			return nil, fmt.Errorf("unknown package name '%s'", name)
		}
	}
	return pkg, nil
}

// ResolvePurl creates a Pkg object from the given purl
// See https://github.com/package-url/purl-spec
func ResolvePurl(purl packageurl.PackageURL) (*pkgmanager.Pkg, error) {
	ecosystem, err := pkgecosystem.ParsePurlType(purl.Type)
	if err != nil {
		return nil, err
	}

	manager := pkgmanager.Manager(ecosystem)
	if manager == nil {
		return nil, pkgecosystem.Unsupported(purl.Type)
	}

	// Prepend package namespace to package name, if present
	var pkgName string
	if purl.Namespace != "" {
		pkgName = purl.Namespace + "/" + purl.Name
	} else {
		pkgName = purl.Name
	}

	// Get the latest package version if not specified in the purl
	pkg, err := ResolvePkg(manager, pkgName, purl.Version, "")
	if err != nil {
		return nil, err
	}

	return pkg, nil
}
