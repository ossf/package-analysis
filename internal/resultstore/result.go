package resultstore

// Pkg describes the various package details used to populate the package part
// of the analysis results.
type Pkg interface {
	Name() string
	Version() string
	Ecosystem() string
}

type pkg struct {
	Name      string
	Version   string
	Ecosystem string
}

type result struct {
	Package          pkg
	CreatedTimestamp int64
	Analysis         interface{}
}
