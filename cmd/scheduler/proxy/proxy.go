package proxy

import (
	"context"
	"log/slog"

	"github.com/ossf/package-analysis/internal/log"
	"gocloud.dev/pubsub"
)

type MessageMutateFunc func(*pubsub.Message) (*pubsub.Message, error)

type PubSubProxy struct {
	topic        *pubsub.Topic
	subscription *pubsub.Subscription
}

func New(topic *pubsub.Topic, subscription *pubsub.Subscription) *PubSubProxy {
	return &PubSubProxy{
		topic:        topic,
		subscription: subscription,
	}
}

func (proxy *PubSubProxy) Listen(ctx context.Context, preprocess MessageMutateFunc) error {
	for {
		msg, err := proxy.subscription.Receive(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "Error receiving message", "error", err)
			return err
		}
		go func(m *pubsub.Message) {
			innerCtx := log.ContextWithAttrs(ctx, slog.String("message_id", m.LoggableID))
			outMsg, err := preprocess(msg)
			if err != nil {
				// Failure to parse and process messages should result in an acknowledgement
				// to avoid the message being redelivered.
				slog.WarnContext(innerCtx, "Error processing message", "error", err)
				m.Ack()
				return
			}
			slog.InfoContext(innerCtx, "Sending message to topic")
			if err := proxy.topic.Send(ctx, outMsg); err != nil {
				slog.ErrorContext(ctx, "Error sending message", "error", err)
				return
			}
			slog.InfoContext(innerCtx, "Sent message successfully")
			msg.Ack()
		}(msg)
	}
}
