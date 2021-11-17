package pkgecosystem

type PkgManager struct {
	Name       string
	CommandFmt func(string, string) string
	GetLatest  func(string) string
	Image      string
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
