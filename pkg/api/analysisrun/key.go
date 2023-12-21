package analysisrun

import (
	"strings"

	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

type Key struct {
	Ecosystem pkgecosystem.Ecosystem `json:"Ecosystem"`
	Name      string                 `json:"Name"`
	Version   string                 `json:"Version"`
}

func (k Key) String() string {
	return strings.Join([]string{string(k.Ecosystem), k.Name, k.Version}, "-")
}
