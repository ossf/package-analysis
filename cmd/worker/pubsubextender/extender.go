package pubsubextender

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/ossf/package-analysis/internal/featureflags"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/gcppubsub"
)

const (
	defaultGracePeriod = 60 * time.Second
	defaultDeadline    = 300 * time.Second
)

var ErrInvalidGracePeriod = errors.New("invalid grace period")

type driver interface {
	// ExtendMessageDeadline sends a request to the pubsub service to change
	// the deadline for the supplied message to the supplied deadline duration.
	ExtendMessageDeadline(ctx context.Context, msg *pubsub.Message, deadline time.Duration) error

	// GetSubscriptionDeadline is used to get the existing deadline period for
	// the pubsub subscription. This deadline is used by default when extending
	// the message deadline.
	GetSubscriptionDeadline(ctx context.Context) (time.Duration, error)
}

type Extender struct {
	driver      driver
	Deadline    time.Duration
	GracePeriod time.Duration
}

func getDriver(u *url.URL, sub *pubsub.Subscription) (driver, error) {
	if !featureflags.PubSubExtender.Enabled() {
		// Use the noopDriver if the feature is disabled.
		return &noopDriver{}, nil
	}

	switch u.Scheme {
	case gcppubsub.Scheme:
		return newGCPDriver(u, sub)
	default:
		return &noopDriver{}, nil
	}
}

func New(ctx context.Context, subURL string, sub *pubsub.Subscription) (*Extender, error) {
	u, err := url.Parse(subURL)
	if err != nil {
		return nil, err
	}

	d, err := getDriver(u, sub)
	if err != nil {
		return nil, err
	}

	deadline, err := d.GetSubscriptionDeadline(ctx)
	if err != nil {
		return nil, err
	}
	if deadline == 0 {
		deadline = defaultDeadline
	}

	return &Extender{
		driver:      d,
		Deadline:    deadline,
		GracePeriod: defaultGracePeriod,
	}, nil
}

type MessageExtender struct {
	ticker   *time.Ticker
	msg      *pubsub.Message
	done     chan bool
	exited   chan error
	callback func()
	running  bool
}

func (e *Extender) Start(ctx context.Context, msg *pubsub.Message, callback func()) (*MessageExtender, error) {
	freq := e.Deadline - e.GracePeriod
	if freq < 0 {
		// GracePeriod is larger than Deadline.
		return nil, fmt.Errorf("%w: deadline %v is smaller than grace period %v", ErrInvalidGracePeriod, e.Deadline, e.GracePeriod)
	}

	me := &MessageExtender{
		ticker:   time.NewTicker(freq),
		msg:      msg,
		done:     make(chan bool),
		exited:   make(chan error),
		callback: callback,
		running:  true,
	}

	go func(ctx context.Context, me *MessageExtender, e *Extender) {
		var err error
		for {
			select {
			case <-me.done:
				me.ticker.Stop()
				me.exited <- err
				return
			case <-me.ticker.C:
				err = e.driver.ExtendMessageDeadline(ctx, me.msg, e.Deadline)
				if err != nil {
					me.ticker.Stop() // don't send any more ticks.
					// at this point the goroutine waits for done so that it
					// can exit cleanly and report the error without race
					// conditions.
				} else if me.callback != nil {
					me.callback()
				}
			}
		}
	}(ctx, me, e)
	return me, nil
}

func (me *MessageExtender) IsRunning() bool {
	return me.running
}

func (me *MessageExtender) Stop() error {
	if me.running {
		// Signal the goroutine that stop has been called and we are done.
		me.done <- true
		// Wait for the goroutine to exit and collect any error that it may
		// have waiting for us.
		err := <-me.exited
		me.running = false
		return err
	}
	return nil
}
