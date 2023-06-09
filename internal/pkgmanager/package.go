package pkgmanager

import (
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

type Pkg struct {
	name    string
	version string
	manager *PkgManager
	local   string
}

func (p *Pkg) Name() string {
	return p.name
}

func (p *Pkg) Version() string {
	return p.version
}

func (p *Pkg) Ecosystem() pkgecosystem.Ecosystem {
	return p.manager.ecosystem
}

func (p *Pkg) EcosystemName() string {
	return string(p.Ecosystem())
}

func (p *Pkg) IsLocal() bool {
	return p.local != ""
}

func (p *Pkg) Manager() *PkgManager {
	return p.manager
}

func (p *Pkg) LocalPath() string {
	return p.local
}
