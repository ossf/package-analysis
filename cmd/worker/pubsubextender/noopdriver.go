package pubsubextender

import (
	"context"
	"time"

	"gocloud.dev/pubsub"
)

type noopDriver struct{}

// ExtendMessageDeadline implements the driver interface.
func (d *noopDriver) ExtendMessageDeadline(ctx context.Context, msg *pubsub.Message, deadline time.Duration) error {
	return nil
}

// GetSubscriptionDeadline implements the driver interface.
func (d *noopDriver) GetSubscriptionDeadline(ctx context.Context) (time.Duration, error) {
	return 0, nil
}
