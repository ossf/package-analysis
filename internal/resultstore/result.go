package resultstore

import (
	"time"
)

// Pkg describes the various package details used to populate the package part
// of the analysis results.
type Pkg interface {
	EcosystemName() string
	Name() string
	Version() string
}

// pkg is an internal representation of a Pkg, which can be marshalled into JSON.
type pkg struct {
	Ecosystem string `json:"Ecosystem"`
	Name      string `json:"Name"`
	Version   string `json:"Version"`
}

// Key is a new version of analysisrun.Key with more fields and is serialised with different JSON keys
type Key struct {
	Ecosystem string    `json:"ecosystem"`
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	Created   time.Time `json:"created"`
}

// DynamicAnalysisRecord is the top-level struct which is serialised to produce JSON results files
// for dynamic analysis
type DynamicAnalysisRecord struct {
	Package          pkg   `json:"Package"`
	CreatedTimestamp int64 `json:"CreatedTimestamp"`
	Analysis         any   `json:"Analysis"`
}

// StaticAnalysisRecord is the top-level struct which is serialised to produce JSON results files
// for static analysis
type StaticAnalysisRecord struct {
	SchemaVersion string `json:"schema_version"`
	Key           Key    `json:"key"`
	Results       any    `json:"results"`
}
