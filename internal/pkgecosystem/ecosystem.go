package pkgecosystem

type PkgManager struct {
	Name      string
	GetLatest func(string) string
	Image     string
}

var (
	SupportedPkgManagers = map[string]PkgManager{
		NPMPackageManager.Name:      NPMPackageManager,
		PyPIPackageManager.Name:     PyPIPackageManager,
		RubyGemsPackageManager.Name: RubyGemsPackageManager,
	}
)

// String implements the Stringer interface to support pretty printing.
func (p PkgManager) String() string {
	return p.Name
}

// Args returns the analysis arguments for the given package.
func (p PkgManager) Args(phase, pkg, ver, local string) []string {
	args := make([]string, 0)

	if local != "" {
		args = append(args, "--local", local)
	} else if ver != "" {
		args = append(args, "--version", ver)
	}

	if phase == "" {
		args = append(args, "all")
	} else {
		args = append(args, phase)
	}

	args = append(args, pkg)

	return args
}
