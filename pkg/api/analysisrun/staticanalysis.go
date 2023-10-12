package analysisrun

import (
	"encoding/json"
	"time"

	"github.com/ossf/package-analysis/internal/staticanalysis/signals"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
	"github.com/ossf/package-analysis/internal/utils/valuecounts"
)

// StaticAnalysisData contains the raw output of a static analysis run, from
// the static analysis sandbox. Its format is not part of the Package Analysis API
// and is subject to change.
type StaticAnalysisData = json.RawMessage

// StaticAnalysisSchemaVersion identifies the static analysis results JSON schema version.
const StaticAnalysisSchemaVersion = "1.0"

// StaticAnalysisRecord is the top-level struct which is serialised to produce static analysis
// JSON files. This struct should not change unless StaticAnalysisSchemaVersion is also incremented.
type StaticAnalysisRecord struct {
	SchemaVersion string                `json:"schema_version"`
	Ecosystem     string                `json:"ecosystem"`
	Name          string                `json:"name"`
	Version       string                `json:"version"`
	Created       time.Time             `json:"created"`
	Results       StaticAnalysisResults `json:"results"`
}

// StaticAnalysisResults holds the output data from static analysis data, as part of the
// StaticAnalysisRecord struct which is a part of the Package Analysis API. These structs
// are serialised to JSON to produce the JSON data files for static analysis.
type StaticAnalysisResults struct {
	Files []StaticAnalysisFileResult `json:"files"`
}

func (r *StaticAnalysisResults) CreateRecord(k Key) *StaticAnalysisRecord {
	return &StaticAnalysisRecord{
		SchemaVersion: StaticAnalysisSchemaVersion,
		Ecosystem:     k.Ecosystem.String(),
		Name:          k.Name,
		Version:       k.Version,
		Created:       time.Now().UTC(),
		Results:       *r,
	}
}

type StaticAnalysisFileResult struct {
	Filename              string                       `json:"filename"`
	DetectedType          string                       `json:"detected_type,omitempty"`
	Size                  int64                        `json:"size,omitempty"`
	Sha256                string                       `json:"sha256,omitempty"`
	LineLengths           valuecounts.ValueCounts      `json:"line_lengths,omitempty"`
	Js                    StaticAnalysisJsData         `json:"js,omitempty"`
	IdentifierLengths     valuecounts.ValueCounts      `json:"identifier_lengths,omitempty"`
	StringLengths         valuecounts.ValueCounts      `json:"string_lengths,omitempty"`
	Base64Strings         []string                     `json:"base64_strings,omitempty"`
	HexStrings            []string                     `json:"hex_strings,omitempty"`
	IPAddresses           []string                     `json:"ip_addresses,omitempty"`
	URLs                  []string                     `json:"urls,omitempty"`
	SuspiciousIdentifiers signals.SuspiciousIdentifier `json:"suspicious_identifiers,omitempty"`
	EscapedStrings        signals.EscapedString        `json:"escaped_strings,omitempty"`
}

type StaticAnalysisJsData struct {
	Identifiers    []token.Identifier `json:"identifiers"`
	StringLiterals []token.String     `json:"string_literals"`
	IntLiterals    []token.Int        `json:"int_literals"`
	FloatLiterals  []token.Float      `json:"float_literals"`
	Comments       []token.Comment    `json:"comments"`
}
