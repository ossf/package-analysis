package notification

import (
	"encoding/json"
	"fmt"
	"gocloud.dev/pubsub"

	"github.com/ossf/package-analysis/pkg/pkgidentifier"
)

// AnalysisCompletion is a struct representing the message sent to notify
// package analysis completion.
type AnalysisCompletion struct {
	Package pkgidentifier.PkgIdentifier
}

// ParseJSON takes in a notification JSON and returns an AnalysisCompletion struct.  
func ParseJSON(msg *pubsub.Message) (AnalysisCompletion, error) {
	notification := AnalysisCompletion{}
	if err := json.Unmarshal(msg.Body, &notification); err != nil {
		return notification, fmt.Errorf("error unmarshalling json: %w", err)
	}
	return notification, nil
}
