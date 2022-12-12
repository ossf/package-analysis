package resultstore

// Pkg describes the various package details used to populate the package part
// of the analysis results.
type Pkg interface {
	Name() string
	Version() string
	EcosystemName() string
}

// A representation of a Pkg, which can be marshalled into JSON
type PkgIdentifier struct {
	Name      string
	Version   string
	Ecosystem string
}

type result struct {
	Package          PkgIdentifier
	CreatedTimestamp int64
	Analysis         interface{}
}
