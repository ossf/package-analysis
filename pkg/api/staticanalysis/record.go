package staticanalysis

import (
	"encoding/json"
	"time"

	"github.com/ossf/package-analysis/pkg/api/analysisrun"
	"github.com/ossf/package-analysis/pkg/api/staticanalysis/token"
	"github.com/ossf/package-analysis/pkg/valuecounts"
)

// SandboxData contains the raw output of a static analysis run, from
// the static analysis sandbox. Its format is not part of the Package Analysis API
// and is subject to change.
type SandboxData = json.RawMessage

// SchemaVersion identifies the static analysis results JSON schema version.
const SchemaVersion = "1.0"

// Record is the top-level struct which is serialised to produce static analysis
// JSON files. This struct should not change unless SchemaVersion is also incremented.
type Record struct {
	SchemaVersion string    `json:"schema_version"`
	Ecosystem     string    `json:"ecosystem"`
	Name          string    `json:"name"`
	Version       string    `json:"version"`
	Created       time.Time `json:"created"`
	Results       Results   `json:"results"`
}

// Results holds the output data from static analysis data, as part of the
// Record struct which is a part of the Package Analysis API. These structs
// are serialised to JSON to produce the JSON data files for static analysis.
type Results struct {
	Files []FileResult `json:"files"`
}

// CreateRecord associates a set of static analysis Results with an identifying Key,
// to produce a Record object that can be serialised.
func CreateRecord(r *Results, k analysisrun.Key) *Record {
	return &Record{
		SchemaVersion: SchemaVersion,
		Ecosystem:     k.Ecosystem.String(),
		Name:          k.Name,
		Version:       k.Version,
		Created:       time.Now().UTC(),
		Results:       *r,
	}
}

// FileResult holds static analysis data for a single file. Filename is the only
// mandatory field, and holds the path to the file relative to the package root.
// Other fields may be present or missing depending on whether relevant data was collected.
type FileResult struct {
	Filename              string                   `json:"filename"`
	DetectedType          string                   `json:"detected_type,omitempty"`
	Size                  int64                    `json:"size,omitempty"`
	SHA256                string                   `json:"sha256,omitempty"`
	LineLengths           *valuecounts.ValueCounts `json:"line_lengths,omitempty"`
	Js                    *JsData                  `json:"js,omitempty"`
	IdentifierLengths     *valuecounts.ValueCounts `json:"identifier_lengths,omitempty"`
	StringLengths         *valuecounts.ValueCounts `json:"string_lengths,omitempty"`
	Base64Strings         []string                 `json:"base64_strings,omitempty"`
	HexStrings            []string                 `json:"hex_strings,omitempty"`
	IPAddresses           []string                 `json:"ip_addresses,omitempty"`
	URLs                  []string                 `json:"urls,omitempty"`
	SuspiciousIdentifiers []SuspiciousIdentifier   `json:"suspicious_identifiers,omitempty"`
	EscapedStrings        []EscapedString          `json:"escaped_strings,omitempty"`
}

type JsData struct {
	Identifiers    []token.Identifier `json:"identifiers"`
	StringLiterals []token.String     `json:"string_literals"`
	IntLiterals    []token.Int        `json:"int_literals"`
	FloatLiterals  []token.Float      `json:"float_literals"`
	Comments       []token.Comment    `json:"comments"`
}
