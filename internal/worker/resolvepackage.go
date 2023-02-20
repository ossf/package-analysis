package worker

import (
	"fmt"

	"github.com/ossf/package-analysis/internal/pkgecosystem"
)

// ResolvePkg creates a Pkg object with the arguments passed to the worker process.
func ResolvePkg(manager *pkgecosystem.PkgManager, name, version, localPath string) (pkg *pkgecosystem.Pkg, err error) {
	switch {
	case localPath != "":
		pkg = manager.Local(name, version, localPath)
	case version != "":
		pkg = manager.Package(name, version)
	default:
		pkg, err = manager.Latest(name)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest version for package %s: %w", name, err)
		}
	}
	return pkg, nil
}
