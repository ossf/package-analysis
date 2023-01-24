package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"gocloud.dev/pubsub"

	"github.com/ossf/package-analysis/pkg/notification"
	"github.com/ossf/package-analysis/pkg/pkgidentifier"
)

func PublishAnalysisCompletion(ctx context.Context, notificationTopic *pubsub.Topic, name, version, ecosystem string) error {
	pkgDetails := pkgidentifier.PkgIdentifier{name, version, ecosystem}
	notificationMsg, err := json.Marshal(notification.AnalysisCompletion{pkgDetails})
	if err != nil {
		return fmt.Errorf("failed to encode completion notification: %w", err)
	}
	err = notificationTopic.Send(ctx, &pubsub.Message{
		Body: []byte(notificationMsg),
		Metadata: nil,
	})
	if err != nil {
		return fmt.Errorf("failed to send completion notification: %w", err)
	}
	return nil
}