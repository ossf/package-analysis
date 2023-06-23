package pubsubextender

import (
	"context"
	"errors"
	"testing"
	"time"

	"gocloud.dev/pubsub"
)

type testDriver struct {
	extendCount int
	lastExtend  struct {
		msg      *pubsub.Message
		deadline time.Duration
	}
	err error
}

// ExtendMessageDeadline implements the driver interface.
func (d *testDriver) ExtendMessageDeadline(ctx context.Context, msg *pubsub.Message, deadline time.Duration) error {
	d.extendCount++
	d.lastExtend.msg = msg
	d.lastExtend.deadline = deadline
	return d.err
}

// GetSubscriptionDeadline implements the driver interface.
func (d *testDriver) GetSubscriptionDeadline(ctx context.Context) (time.Duration, error) {
	return 0, nil
}

func TestNew(t *testing.T) {
	e, err := New(context.Background(), "not://a/real/pubsub/subscription", nil)
	if err != nil {
		t.Fatalf("New() = %v; want no error", err)
	}
	if e.Deadline != defaultDeadline {
		t.Errorf("Deadline = %v; want %v", e.Deadline, defaultDeadline)
	}
	if e.GracePeriod != defaultGracePeriod {
		t.Errorf("GracePeriod = %v; want %v", e.GracePeriod, defaultGracePeriod)
	}
}

func TestExtender(t *testing.T) {
	wantDeadline := 123 * time.Millisecond
	d := &testDriver{}
	e := &Extender{
		driver:      d,
		Deadline:    wantDeadline,
		GracePeriod: 50 * time.Millisecond,
	}
	ctx := context.Background()
	wantMsg := &pubsub.Message{
		LoggableID: "test-msg-0",
		Metadata:   map[string]string{"test": "test"},
	}

	me, err := e.Start(ctx, wantMsg, func() {
		if got := d.lastExtend.msg; got != wantMsg {
			t.Errorf("ExtendMessageDeadline got message %v, want %v", got, wantMsg)
		}
		if got := d.lastExtend.deadline; got != wantDeadline {
			t.Errorf("ExtendMessageDeadline got deadline %v, want %v", got, wantDeadline)
		}

		d.lastExtend.msg = nil
		d.lastExtend.deadline = 0
	})
	if err != nil {
		t.Fatalf("Start() = %v; want no error", err)
	}
	if !me.IsRunning() {
		t.Errorf("IsRunning()#1 = false; want true")
	}

	time.Sleep(500 * time.Millisecond)

	if err := me.Stop(); err != nil {
		t.Errorf("Err() = %v; want nil", err)
	}

	if d.extendCount == 0 {
		t.Errorf("ExtendMessageDeadline never called")
	}
	if me.IsRunning() {
		t.Errorf("IsRunning()#2 = true; want false")
	}

	// Reset so we can ensure that it has stoppped.
	d.extendCount = 0

	time.Sleep(500 * time.Millisecond)
	if d.extendCount != 0 {
		t.Errorf("Extender not stopped")
	}

	// Calling stop again does nothing and has no error.
	if err := me.Stop(); err != nil {
		t.Errorf("Err() = %v; want nil", err)
	}
}

func TestExtender_Error(t *testing.T) {
	wantErr := errors.New("failed")
	d := &testDriver{
		err: wantErr,
	}
	e := &Extender{
		driver:      d,
		Deadline:    100 * time.Millisecond,
		GracePeriod: 50 * time.Millisecond,
	}
	ctx := context.Background()
	wantMsg := &pubsub.Message{
		LoggableID: "test-msg-0",
		Metadata:   map[string]string{"test": "test"},
	}

	me, err := e.Start(ctx, wantMsg, nil)
	if err != nil {
		t.Fatalf("Start() = %v; want no error", err)
	}

	time.Sleep(500 * time.Millisecond)

	if err := me.Stop(); err == nil {
		t.Errorf("Err() = %v; want %v", err, wantErr)
	}
	if d.extendCount != 1 {
		t.Errorf("ExtendMessageDeadline called after error")
	}
	// Calling stop again does nothing and has no error.
	if err := me.Stop(); err != nil {
		t.Errorf("Err() = %v; want nil", err)
	}
}
