package pkgecosystem

type Pkg struct {
	name    string
	version string
	manager *PkgManager
	local   string
	command string
}

func (p *Pkg) Name() string {
	return p.name
}

func (p *Pkg) Version() string {
	return p.version
}

func (p *Pkg) Ecosystem() string {
	return p.manager.name
}

func (p *Pkg) IsLocal() bool {
	return p.local != ""
}

func (p *Pkg) Manager() *PkgManager {
	return p.manager
}

// Command returns the analysis command for the package.
func (p *Pkg) Command(phase RunPhase) []string {
	args := make([]string, 0)
	args = append(args, p.manager.command)

	if p.local != "" {
		args = append(args, "--local", p.local)
	} else if p.version != "" {
		args = append(args, "--version", p.version)
	}

	if phase == "" {
		args = append(args, "all")
	} else {
		args = append(args, string(phase))
	}

	args = append(args, p.name)

	return args
}
