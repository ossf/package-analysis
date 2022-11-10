package pkgecosystem

import (
	"strings"
)

// RunPhase
// Represents a way to 'run' a package during its usage lifecycle
// This is relevant to dynamic analysis
type RunPhase string

const (
	Import  RunPhase = "import"
	Install RunPhase = "install"
)

// PkgManager
// Represents how packages from a common ecosystem are accessed
type PkgManager struct {
	name      string
	image     string
	command   string
	latest    func(string) (string, error)
	runPhases []RunPhase
}

var (
	supportedPkgManagers = map[string]*PkgManager{
		npmPkgManager.name:       &npmPkgManager,
		pypiPkgManager.name:      &pypiPkgManager,
		rubygemsPkgManager.name:  &rubygemsPkgManager,
		packagistPkgManager.name: &packagistPkgManager,
		cratesPkgManager.name:    &cratesPkgManager,
	}
)

func Manager(ecosystem string) *PkgManager {
	return supportedPkgManagers[ecosystem]
}

// String implements the Stringer interface to support pretty printing.
func (p *PkgManager) String() string {
	return p.name
}

func (p *PkgManager) DynamicAnalysisImage() string {
	return p.image
}

func (p *PkgManager) RunPhases() []RunPhase {
	return p.runPhases
}

func (p *PkgManager) Latest(name string) (*Pkg, error) {
	name = normalizePkgName(name)
	version, err := p.latest(name)
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

func normalizePkgName(pkg string) string {
	return strings.ToLower(pkg)
}
