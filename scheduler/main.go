package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jordan-wright/ossmalware/pkg/library"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
)

const (
	analysisImage = "gcr.io/ossf-malware-analysis/python"
)

var supportedPkgManagers = map[string]bool{
	"npm":      true,
	"pypi":     true,
	"rubygems": true,
}

func main() {
	subscriptionURL := os.Getenv("OSSMALWARE_SUBSCRIPTION_URL")
	ctx := context.Background()
	sub, err := pubsub.OpenSubscription(ctx, subscriptionURL)
	if err != nil {
		panic(err)
	}

	topicURL := os.Getenv("OSSMALWARE_WORKER_TOPIC")
	topic, err := pubsub.OpenTopic(ctx, topicURL)
	if err != nil {
		panic(err)
	}

	for {
		msg, err := sub.Receive(ctx)
		if err != nil {
			log.Println("error receiving message: ", err)
			continue
		}
		go func(m *pubsub.Message) {
			log.Println("handling message: ", string(m.Body))
			pkg := library.Package{}
			if err := json.Unmarshal(m.Body, &pkg); err != nil {
				log.Println("error unmarshalling json: ", err)
				return
			}
			if err := handlePkg(ctx, topic, pkg); err != nil {
				fmt.Println("Error: ", err)
				msg.Nack()
				return
			}
			msg.Ack()
		}(msg)
	}
}

func handlePkg(ctx context.Context, topic *pubsub.Topic, pkg library.Package) error {
	if _, ok := supportedPkgManagers[pkg.Type]; !ok {
		log.Println("unknown package type: ", pkg.Type)
		return nil
	}

	return topic.Send(ctx, &pubsub.Message{
		Metadata: map[string]string{
			"name":      pkg.Name,
			"ecosystem": pkg.Type,
			"version":   pkg.Version,
		},
	})
}
