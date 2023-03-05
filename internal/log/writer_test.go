package log_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"golang.org/x/exp/slices"

	"github.com/ossf/package-analysis/internal/log"
)

func initLogs(t *testing.T, l zapcore.Level) (*zap.Logger, *observer.ObservedLogs) {
	t.Helper()
	core, obs := observer.New(l)
	return zap.New(core), obs
}

func TestNewWriter_SingleLine(t *testing.T) {
	logger, obs := initLogs(t, zapcore.DebugLevel)
	w := log.NewWriter(logger, zapcore.InfoLevel)

	want := "this is the log message"

	_, err := io.Copy(w, bytes.NewBuffer([]byte(want)))
	w.Close()

	if err != nil {
		t.Fatalf("Writing failed: %v", err)
	}
	if got := obs.Len(); got != 1 {
		t.Fatalf("Got %d log entries; want 1", got)
	}
	entry := obs.All()[0]
	if got := entry.Message; got != want {
		t.Errorf("Got %v entry; want %v", got, want)
	}
}

func TestNewWriter_MultiLine(t *testing.T) {
	logger, obs := initLogs(t, zapcore.DebugLevel)
	w := log.NewWriter(logger, zapcore.InfoLevel)

	want := []string{
		"one",
		"two",
		"three",
		"four",
	}

	_, err := io.Copy(w, bytes.NewBuffer([]byte(strings.Join(want, "\n"))))
	w.Close()

	if err != nil {
		t.Fatalf("Writing failed: %v", err)
	}

	var got []string
	for _, entry := range obs.All() {
		got = append(got, entry.Message)
	}
	if !slices.Equal(got, want) {
		t.Errorf("Got log entries = %v; want %v", got, want)
	}
}

func TestNewWriter_LevelSuppress(t *testing.T) {
	logger, obs := initLogs(t, zapcore.WarnLevel)
	w := log.NewWriter(logger, zapcore.InfoLevel)

	want := "this is the log message"

	_, err := io.Copy(w, bytes.NewBuffer([]byte(want)))
	w.Close()

	if err != nil {
		t.Fatalf("Writing failed: %v", err)
	}
	if got := obs.Len(); got != 0 {
		t.Fatalf("Got %d log entries; want none", got)
	}
}

func TestNewWriter_MultiWithEmptyLine(t *testing.T) {
	logger, obs := initLogs(t, zapcore.DebugLevel)
	w := log.NewWriter(logger, zapcore.InfoLevel)

	in := []string{"one", "two", "", "four"}
	want := []string{"one", "two", "four"}

	_, err := io.Copy(w, bytes.NewBuffer([]byte(strings.Join(in, "\n"))))
	w.Close()

	if err != nil {
		t.Fatalf("Writing failed: %v", err)
	}

	var got []string
	for _, entry := range obs.All() {
		got = append(got, entry.Message)
	}
	if !slices.Equal(got, want) {
		t.Errorf("Got log entries = %v; want %v", got, want)
	}
}

func TestNewWriter_MultiWithTrailingSpaces(t *testing.T) {
	logger, obs := initLogs(t, zapcore.DebugLevel)
	w := log.NewWriter(logger, zapcore.InfoLevel)

	in := []string{"one    ", "two \t \f \v \r", "\t\t\t\t", "four"}
	want := []string{"one", "two", "four"}

	_, err := io.Copy(w, bytes.NewBuffer([]byte(strings.Join(in, "\n"))))
	w.Close()

	if err != nil {
		t.Fatalf("Writing failed: %v", err)
	}

	var got []string
	for _, entry := range obs.All() {
		got = append(got, entry.Message)
	}
	if !slices.Equal(got, want) {
		t.Errorf("Got log entries = %v; want %v", got, want)
	}
}

func TestNewWriter_Empty(t *testing.T) {
	logger, obs := initLogs(t, zapcore.DebugLevel)
	w := log.NewWriter(logger, zapcore.InfoLevel)

	_, err := io.Copy(w, &bytes.Buffer{})
	w.Close()

	if err != nil {
		t.Fatalf("Writing failed: %v", err)
	}
	if got := obs.Len(); got != 0 {
		t.Fatalf("Got %d log entries; want none", got)
	}
}

func TestNewWriter_TrailingNewline(t *testing.T) {
	logger, obs := initLogs(t, zapcore.DebugLevel)
	w := log.NewWriter(logger, zapcore.InfoLevel)

	want := "this is the log message"

	_, err := io.Copy(w, bytes.NewBuffer([]byte(want+"\n")))
	w.Close()

	if err != nil {
		t.Fatalf("Writing failed: %v", err)
	}
	if got := obs.Len(); got != 1 {
		t.Fatalf("Got %d log entries; want 1", got)
	}
	entry := obs.All()[0]
	if got := entry.Message; got != want {
		t.Errorf("Got %v entry; want %v", got, want)
	}
}

func TestNewWriter_MultiWrites(t *testing.T) {
	logger, obs := initLogs(t, zapcore.DebugLevel)
	w := log.NewWriter(logger, zapcore.InfoLevel)

	in1 := []string{"one", "two", "", "fourty "}
	in2 := []string{"two", "...", "done"}
	want := []string{"one", "two", "fourty two", "...", "done"}

	_, err := io.Copy(w, bytes.NewBuffer([]byte(strings.Join(in1, "\n"))))
	if err != nil {
		t.Fatalf("Writing #1 failed: %v", err)
	}

	_, err = io.Copy(w, bytes.NewBuffer([]byte(strings.Join(in2, "\n"))))
	if err != nil {
		t.Fatalf("Writing #2 failed: %v", err)
	}
	w.Close()

	var got []string
	for _, entry := range obs.All() {
		got = append(got, entry.Message)
	}
	if !slices.Equal(got, want) {
		t.Errorf("Got log entries = %v; want %v", got, want)
	}
}
