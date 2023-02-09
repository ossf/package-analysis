package pkgidentifier

import (
	"strings"
)

type PkgIdentifier struct {
	Name		string
	Version 	string
	Ecosystem	string
}

func (pkg PkgIdentifier) String() string {
	return strings.Join([]string{pkg.Ecosystem, pkg.Name, pkg.Version}, "-")
}
