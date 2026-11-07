package main

import (
	"context"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestShutdownRequested_LiveContext(t *testing.T) {
	ctx := context.Background()
	if shutdownRequested(ctx) {
		t.Errorf("shutdownRequested() = true for a live context, want false")
	}
}

func TestShutdownRequested_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !shutdownRequested(ctx) {
		t.Errorf("shutdownRequested() = false for a cancelled context, want true")
	}
}

func TestShutdownRequested_OnSIGTERM(t *testing.T) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	if shutdownRequested(ctx) {
		t.Fatalf("shutdownRequested() = true before SIGTERM, want false")
	}

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM to self: %v", err)
	}

	// The signal is delivered asynchronously; wait briefly for the context to
	// observe it before asserting.
	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("context was not cancelled within 2s of SIGTERM")
	}

	if !shutdownRequested(ctx) {
		t.Errorf("shutdownRequested() = false after SIGTERM, want true")
	}
}
