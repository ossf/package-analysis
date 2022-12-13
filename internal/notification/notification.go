package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"gocloud.dev/pubsub"
)

type pkgIdentifier struct {
	Name      string
	Version   string
	Ecosystem string
}

type NotificationMessage struct {
	Package	pkgIdentifier
}

func PublishNotification(ctx context.Context, notificationTopic *pubsub.Topic, name, version, ecosystem string) error {
	pkgDetails := pkgIdentifier{name, version, ecosystem}
	notificationMsg, err := json.Marshal(NotificationMessage{pkgDetails})
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