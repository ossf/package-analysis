package log_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"golang.org/x/exp/slices"

	"github.com/ossf/package-analysis/internal/log"
)

func initLogs(t *testing.T) (*slog.Logger, *testHandler) {
	t.Helper()
	h := &testHandler{}
	return slog.New(h), h
}

func TestNewWriter_SingleLine(t *testing.T) {
	logger, h := initLogs(t)
	w := log.NewWriter(context.Background(), logger, slog.LevelInfo)

	want := "this is the log message"
	wantLevel := slog.LevelInfo

	_, err := io.Copy(w, bytes.NewBuffer([]byte(want)))
	w.Close()

	if err != nil {
		t.Fatalf("Writing failed: %v", err)
	}
	if got := h.Len(); got != 1 {
		t.Fatalf("Got %d log records; want 1", got)
	}
	r := h.All()[0]
	if got := r.Message; got != want {
		t.Errorf("Got %v message; want %v", got, want)
	}
	if got := r.Level; got != wantLevel {
		t.Errorf("Get %v level; want %v", got, want)
	}
}

func TestNewWriter_MultiLine(t *testing.T) {
	logger, h := initLogs(t)
	w := log.NewWriter(context.Background(), logger, slog.LevelInfo)

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
	for _, r := range h.All() {
		got = append(got, r.Message)
	}
	if !slices.Equal(got, want) {
		t.Errorf("Got log records = %v; want %v", got, want)
	}
}

func TestNewWriter_MultiWithEmptyLine(t *testing.T) {
	logger, h := initLogs(t)
	w := log.NewWriter(context.Background(), logger, slog.LevelInfo)

	in := []string{"one", "two", "", "four"}
	want := []string{"one", "two", "four"}

	_, err := io.Copy(w, bytes.NewBuffer([]byte(strings.Join(in, "\n"))))
	w.Close()

	if err != nil {
		t.Fatalf("Writing failed: %v", err)
	}

	var got []string
	for _, r := range h.All() {
		got = append(got, r.Message)
	}
	if !slices.Equal(got, want) {
		t.Errorf("Got log records = %v; want %v", got, want)
	}
}

func TestNewWriter_MultiWithTrailingSpaces(t *testing.T) {
	logger, h := initLogs(t)
	w := log.NewWriter(context.Background(), logger, slog.LevelInfo)

	in := []string{"one    ", "two \t \f \v \r", "\t\t\t\t", "four"}
	want := []string{"one", "two", "four"}

	_, err := io.Copy(w, bytes.NewBuffer([]byte(strings.Join(in, "\n"))))
	w.Close()

	if err != nil {
		t.Fatalf("Writing failed: %v", err)
	}

	var got []string
	for _, r := range h.All() {
		got = append(got, r.Message)
	}
	if !slices.Equal(got, want) {
		t.Errorf("Got log records = %v; want %v", got, want)
	}
}

func TestNewWriter_Empty(t *testing.T) {
	logger, h := initLogs(t)
	w := log.NewWriter(context.Background(), logger, slog.LevelInfo)

	_, err := io.Copy(w, &bytes.Buffer{})
	w.Close()

	if err != nil {
		t.Fatalf("Writing failed: %v", err)
	}
	if got := h.Len(); got != 0 {
		t.Fatalf("Got %d log records; want none", got)
	}
}

func TestNewWriter_TrailingNewline(t *testing.T) {
	logger, h := initLogs(t)
	w := log.NewWriter(context.Background(), logger, slog.LevelInfo)

	want := "this is the log message"

	_, err := io.Copy(w, bytes.NewBuffer([]byte(want+"\n")))
	w.Close()

	if err != nil {
		t.Fatalf("Writing failed: %v", err)
	}
	if got := h.Len(); got != 1 {
		t.Fatalf("Got %d log records; want 1", got)
	}
	r := h.All()[0]
	if got := r.Message; got != want {
		t.Errorf("Got %v message; want %v", got, want)
	}
}

func TestNewWriter_MultiWrites(t *testing.T) {
	logger, h := initLogs(t)
	w := log.NewWriter(context.Background(), logger, slog.LevelInfo)

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
	for _, r := range h.All() {
		got = append(got, r.Message)
	}
	if !slices.Equal(got, want) {
		t.Errorf("Got log records = %v; want %v", got, want)
	}
}
