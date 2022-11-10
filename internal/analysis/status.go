package analysis

import (
	"encoding/json"

	"github.com/ossf/package-analysis/internal/sandbox"
)

type Status string

const (
	// StatusCompleted indicates that the analysis run completed successfully.
	StatusCompleted = Status("completed")

	// StatusErrorTimeout indicates that the analysis was aborted due to a
	// timeout.
	StatusErrorTimeout = Status("error_timeout")

	// StatusErrorAnalysis indicates that the package being analyzed failed
	// while running the specified command.
	//
	// The Stdout and Stderr in the Result should be consulted to understand
	// further why it failed.
	StatusErrorAnalysis = Status("error_analysis")

	// StatusErrorOther indicates an error during some part of the analysis
	// excluding errors covered by other statuses.
	StatusErrorOther = Status("error_other")
)

// MarshalJSON implements the json.Marshaler interface.
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func StatusForRunResult(r *sandbox.RunResult) Status {
	switch r.Status() {
	case sandbox.RunStatusSuccess:
		return StatusCompleted
	case sandbox.RunStatusFailure:
		return StatusErrorAnalysis
	case sandbox.RunStatusTimeout:
		return StatusErrorTimeout
	default:
		return StatusErrorOther
	}
}
