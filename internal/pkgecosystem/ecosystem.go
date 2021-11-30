package pkgecosystem

import "strings"

type PkgManager struct {
	Name        string
	CommandPath string
	GetLatest   func(string) string
	Image       string
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

// escape ensures that args being used are safely escaped for usage.
func escape(s string) string {
	if len(s) == 0 {
		return "''"
	}

	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// Command returns the analysis command for the given package.
func (p PkgManager) Command(phase, pkg, ver, local string) string {
	args := make([]string, 0)
	args = append(args, p.CommandPath)

	if local != "" {
		args = append(args, "--local", escape(local))
	} else if ver != "" {
		args = append(args, "--version", escape(ver))
	}

	if phase == "" {
		args = append(args, "all")
	} else {
		args = append(args, phase)
	}

	args = append(args, escape(pkg))

	return strings.Join(args, " ")
}
