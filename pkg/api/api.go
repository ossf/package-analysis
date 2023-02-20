// The api package defines types and some constants used in the external API for Package Analysis.
package api

// Ecosystem represents an open source package ecosystem from which packages can be downloaded.
type Ecosystem string

// RunPhase represents a way to 'run' a package during its usage lifecycle.
// This is relevant to dynamic analysis.
type RunPhase string

const (
	RunPhaseImport  RunPhase = "import"
	RunPhaseInstall RunPhase = "install"

	EcosystemCratesIO  Ecosystem = "crates.io"
	EcosystemNPM       Ecosystem = "npm"
	EcosystemPackagist Ecosystem = "packagist"
	EcosystemPyPI      Ecosystem = "pypi"
	EcosystemRubyGems  Ecosystem = "rubygems"
)
