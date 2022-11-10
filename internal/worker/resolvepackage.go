package worker

import (
	"fmt"

	"github.com/ossf/package-analysis/internal/pkgecosystem"
)

// ResolvePkg creates a Pkg object with the arguments passed to the worker process
func ResolvePkg(manager *pkgecosystem.PkgManager, name, version, localPath string) (pkg *pkgecosystem.Pkg, err error) {
	if localPath != "" {
		pkg = manager.Local(name, version, localPath)
	} else if version != "" {
		pkg = manager.Package(name, version)
	} else {
		pkg, err = manager.Latest(name)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest version for package %s: %v", name, err)
		}
	}
	return pkg, nil
}
