package analysisrun

import (
	"strings"

	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

type Key struct {
	Ecosystem pkgecosystem.Ecosystem
	Name      string
	Version   string
}

func (k Key) String() string {
	return strings.Join([]string{string(k.Ecosystem), k.Name, k.Version}, "-")
}
