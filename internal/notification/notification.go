package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"gocloud.dev/pubsub"

	"github.com/ossf/package-analysis/pkg/api/analysisrun"
	"github.com/ossf/package-analysis/pkg/api/notification"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

func PublishAnalysisCompletion(ctx context.Context, notificationTopic *pubsub.Topic, name, version string, ecosystem pkgecosystem.Ecosystem) error {
	k := analysisrun.Key{Name: name, Version: version, Ecosystem: ecosystem}
	notificationMsg, err := json.Marshal(notification.AnalysisRunComplete{Key: k})
	if err != nil {
		return fmt.Errorf("failed to encode completion notification: %w", err)
	}
	err = notificationTopic.Send(ctx, &pubsub.Message{
		Body:     notificationMsg,
		Metadata: nil,
	})
	if err != nil {
		return fmt.Errorf("failed to send completion notification: %w", err)
	}
	return nil
}
