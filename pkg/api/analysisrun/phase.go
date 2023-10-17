package analysisrun

// DynamicPhase represents a way to 'run' a package during its usage lifecycle.
// This is relevant to dynamic analysis.
type DynamicPhase string

const (
	DynamicPhaseExecute DynamicPhase = "execute"
	DynamicPhaseImport  DynamicPhase = "import"
	DynamicPhaseInstall DynamicPhase = "install"
)

func DefaultDynamicPhases() []DynamicPhase {
	// ordered in sequence of operation
	return []DynamicPhase{DynamicPhaseInstall, DynamicPhaseImport, DynamicPhaseExecute}
}
