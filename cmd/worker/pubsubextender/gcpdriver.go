package pubsubextender

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	api "cloud.google.com/go/pubsub/apiv1"
	pb "cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/gcppubsub"
)

const (
	gcpMinAckDeadline = 10 * time.Second
	gcpMaxAckDeadline = 600 * time.Second
)

var subscriptionPathRE = regexp.MustCompile("^projects/.+/subscriptions/.+$")

type gcpDriver struct {
	client *api.SubscriberClient
	path   string
}

func newGCPDriver(u *url.URL, sub *pubsub.Subscription) (driver, error) {
	d := &gcpDriver{}

	if u.Scheme != gcppubsub.Scheme {
		return nil, errors.New("unsupported scheme")
	}

	subPath := path.Join(u.Host, u.Path)
	if !subscriptionPathRE.MatchString(subPath) {
		// assume the Host is Project ID and Path is the subscription
		subPath = fmt.Sprintf("projects/%s/subscriptions/%s", u.Host, strings.TrimPrefix(u.Path, "/"))
	}

	var c *api.SubscriberClient
	if !sub.As(&c) {
		return nil, errors.New("not a GCP subscription")
	}
	d.client = c
	d.path = subPath
	return d, nil
}

// ExtendMessageDeadline implements the driver interface.
func (d *gcpDriver) ExtendMessageDeadline(ctx context.Context, msg *pubsub.Message, deadline time.Duration) error {
	// Ensure the deadline is within acceptable bounds.
	if deadline < gcpMinAckDeadline {
		deadline = gcpMinAckDeadline
	} else if deadline > gcpMaxAckDeadline {
		deadline = gcpMaxAckDeadline
	}

	var rm *pb.ReceivedMessage
	if !msg.As(&rm) {
		return errors.New("not a gcp message")
	}

	if err := d.client.ModifyAckDeadline(ctx, &pb.ModifyAckDeadlineRequest{
		Subscription:       d.path,
		AckIds:             []string{rm.AckId},
		AckDeadlineSeconds: int32(deadline / time.Second),
	}); err != nil {
		return fmt.Errorf("failed to extend message deadline: %w", err)
	}

	return nil
}

// GetSubscriptionDeadline implements the driver interface.
func (d *gcpDriver) GetSubscriptionDeadline(ctx context.Context) (time.Duration, error) {
	resp, err := d.client.GetSubscription(ctx, &pb.GetSubscriptionRequest{Subscription: d.path})
	if err != nil {
		return 0, err
	}
	return time.Duration(resp.GetAckDeadlineSeconds()) * time.Second, nil
}
