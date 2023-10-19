package analysisrun

// DynamicPhase represents a way to 'run' a package during its usage lifecycle.
// This is relevant to dynamic analysis.
type DynamicPhase string

const (
	DynamicPhaseExecute DynamicPhase = "execute"
	DynamicPhaseImport  DynamicPhase = "import"
	DynamicPhaseInstall DynamicPhase = "install"
)

// DefaultDynamicPhases the subset of AllDynamicPhases that are supported
// by every ecosystem, and are run by default for dynamic analysis.
func DefaultDynamicPhases() []DynamicPhase {
	return []DynamicPhase{DynamicPhaseInstall, DynamicPhaseImport}
}

// AllDynamicPhases lists each phase of dynamic analysis in order
// that they are run. Each phase depends on the previous phases.
func AllDynamicPhases() []DynamicPhase {
	return []DynamicPhase{DynamicPhaseInstall, DynamicPhaseImport, DynamicPhaseExecute}
}
