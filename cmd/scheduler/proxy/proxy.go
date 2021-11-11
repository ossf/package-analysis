package proxy

import (
	"context"
	"fmt"
	"log"

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
			log.Println("Error receiving message: ", err)
			return err
		}
		go func(m *pubsub.Message) {
			outMsg, err := preprocess(msg)
			if err != nil {
				// Failure to parse and process messages should result in an acknowledgement to avoid the message being redelivered.
				log.Println("Error processing message: ", err)
				m.Ack()
				return
			}
			fmt.Println("Sending message to topic")
			if err := proxy.topic.Send(ctx, outMsg); err != nil {
				fmt.Println("Error: ", err)
				return
			}
			fmt.Println("Sent message successfully")
			msg.Ack()
		}(msg)
	}
}
