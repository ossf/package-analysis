package notification

import (
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
)

// AnalysisRunComplete is a struct representing the message sent to notify when
// a package analysis run is complete.
type AnalysisRunComplete struct {
	Key analysisrun.Key
}
