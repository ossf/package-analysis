package resultstore

import "github.com/ossf/package-analysis/pkg/api/pkgecosystem"

// Pkg describes the various package details used to populate the package part
// of the analysis results.
type Pkg interface {
	Ecosystem() pkgecosystem.Ecosystem
	EcosystemName() string
	Name() string
	Version() string
}
