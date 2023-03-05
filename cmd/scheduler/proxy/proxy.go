package proxy

import (
	"context"

	"go.uber.org/zap"
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

func (proxy *PubSubProxy) Listen(ctx context.Context, logger *zap.Logger, preprocess MessageMutateFunc) error {
	for {
		msg, err := proxy.subscription.Receive(ctx)
		if err != nil {
			logger.With(zap.Error(err)).Error("Error receiving message")
			return err
		}
		go func(m *pubsub.Message) {
			logger := logger.With(zap.String("message_id", m.LoggableID))
			outMsg, err := preprocess(msg)
			if err != nil {
				// Failure to parse and process messages should result in an acknowledgement
				// to avoid the message being redelivered.
				logger.With(zap.Error(err)).Warn("Error processing message")
				m.Ack()
				return
			}
			logger.Info("Sending message to topic")
			if err := proxy.topic.Send(ctx, outMsg); err != nil {
				logger.With(zap.Error(err)).Error("Error sending message")
				return
			}
			logger.Info("Sent message successfully")
			msg.Ack()
		}(msg)
	}
}
