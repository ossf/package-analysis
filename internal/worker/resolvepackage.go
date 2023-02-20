package worker

import (
	"fmt"

	"github.com/ossf/package-analysis/internal/pkgmanager"
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
			return nil, fmt.Errorf("failed to get latest version for package %s: %w", name, err)
		}
	}
	return pkg, nil
}
