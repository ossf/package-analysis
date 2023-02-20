package notification

import (
	"github.com/ossf/package-analysis/pkg/pkgidentifier"
)

// AnalysisCompletion is a struct representing the message sent to notify
// package analysis completion.
type AnalysisCompletion struct {
	Package pkgidentifier.PkgIdentifier
}
