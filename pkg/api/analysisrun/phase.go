package analysisrun

// DynamicPhase represents a way to 'run' a package during its usage lifecycle.
// This is relevant to dynamic analysis.
type DynamicPhase string

const (
	DynamicPhaseImport  DynamicPhase = "import"
	DynamicPhaseInstall DynamicPhase = "install"
)

func DefaultDynamicPhases() []DynamicPhase {
	return []DynamicPhase{DynamicPhaseInstall, DynamicPhaseImport}
}
