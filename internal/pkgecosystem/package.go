package pkgecosystem

import "fmt"

// Package provides an interface for working with packages from any supported package ecosystem
type Package interface {
	// Name returns the name of this package
	Name() string
	// Ecosystem returns the ecosystem of this package. It is returned as a string for compatibility reasons,
	// but is always a valid Ecosystem.
	Ecosystem() string
	// Version returns the version of this package
	Version() string
	// LocalPath specifies a path to a local archive associated with this package, if one exists.
	LocalPath() string
	// Command returns the full command (with arguments) for dynamic analysis
	Command(phase RunPhase) []string
	// Download attempts to download an archive for the given package name and version and returns the archive path
	Download() (string, error)
	// DynamicAnalysisImage returns the name of container image to use for dynamic analysis
	DynamicAnalysisImage() string
	// DynamicRunPhases returns a list of phases to run for dynamic analysis
	DynamicRunPhases() []RunPhase
	// String formats this Package for pretty printing
	String() string
	// baseCommand returns the (incomplete) base command to use for dynamic analysis
	baseCommand() string
}

func notImplemented() {
	panic("Not implemented")
}

// packageToString makes a pretty string representation of the given Package
func packageToString(p Package) string {
	return fmt.Sprintf("%s package: %s, version %s", p.Ecosystem(), p.Name(), p.Version())
}

// phaseCommand returns the full analysis command for the package.
func phaseCommand(p Package, phase RunPhase) []string {
	args := make([]string, 0)
	args = append(args, p.baseCommand())

	if p.LocalPath() != "" {
		args = append(args, "--local", p.LocalPath())
	} else if p.Version() != "" {
		args = append(args, "--version", p.Version())
	}

	if phase == "" {
		args = append(args, "all")
	} else {
		args = append(args, string(phase))
	}

	args = append(args, p.Name())

	return args
}
