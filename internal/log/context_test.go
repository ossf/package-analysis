package log_test

import (
	"context"
	"testing"

	"log/slog"

	"github.com/ossf/package-analysis/internal/log"
)

func assertRecordAttrs(t *testing.T, r slog.Record, attrs []slog.Attr) {
	t.Helper()

	wantLen := len(attrs)
	gotLen := r.NumAttrs()
	if wantLen != gotLen {
		t.Errorf("record.NumAttrs() = %v; want %v", gotLen, wantLen)
	}

	r.Attrs(func(a slog.Attr) bool {
		for _, attr := range attrs {
			if a.Equal(attr) {
				return true
			}
		}
		t.Errorf("unexpected attr %v", a)
		return true
	})
}

func TestContextWithAttrs(t *testing.T) {
	attr1 := slog.Any("hello", "world")
	attr2 := slog.Int("meaning", 42)
	attr3 := slog.String("a", "b")

	h := &testHandler{}
	logger := slog.New(log.NewContextLogHandler(h))

	ctx := context.Background()

	// Add attrs to the context and ensure they are used.
	ctx = log.ContextWithAttrs(ctx, attr1, attr2)
	logger.InfoContext(ctx, "test", "a", "b")
	assertRecordAttrs(t, h.LastRecord(), []slog.Attr{attr1, attr2, attr3})
}

func TestContextWithAttrs_InnerCtx(t *testing.T) {
	attr1 := slog.Any("hello", "world")
	attr2 := slog.Int("meaning", 42)
	attr3 := slog.Any("complex", struct{ a string }{a: "string"})

	h := &testHandler{}
	logger := slog.New(log.NewContextLogHandler(h))

	ctx := context.Background()
	ctx = log.ContextWithAttrs(ctx, attr1, attr2)

	// Add more attrs to the context and ensure they are used.
	innerCtx := log.ContextWithAttrs(ctx, attr3)
	logger.InfoContext(innerCtx, "test")
	assertRecordAttrs(t, h.LastRecord(), []slog.Attr{attr1, attr2, attr3})
}

func TestContextWithAttrs_OuterAfterInnerCtx(t *testing.T) {
	attr1 := slog.Any("hello", "world")
	attr2 := slog.Int("meaning", 42)
	attr3 := slog.Any("complex", struct{ a string }{a: "string"})

	h := &testHandler{}
	logger := slog.New(log.NewContextLogHandler(h))

	ctx := context.Background()
	ctx = log.ContextWithAttrs(ctx, attr1, attr2)
	_ = log.ContextWithAttrs(ctx, attr3)

	// Use the earlier context to ensure the innerCtx attrs are not included.
	logger.InfoContext(ctx, "test")
	assertRecordAttrs(t, h.LastRecord(), []slog.Attr{attr1, attr2})
}

func TestContextWithAttrs_NoAttrs(t *testing.T) {
	attr1 := slog.String("a", "b")

	h := &testHandler{}
	logger := slog.New(log.NewContextLogHandler(h))

	ctx := context.Background()
	ctx = log.ContextWithAttrs(ctx)

	logger.InfoContext(ctx, "test", "a", "b")
	assertRecordAttrs(t, h.LastRecord(), []slog.Attr{attr1})
}

func TestLoggerWithContext(t *testing.T) {
	attr1 := slog.Any("hello", "world")
	attr2 := slog.Int("meaning", 42)
	attr3 := slog.String("a", "b")

	h := &testHandler{}
	logger := slog.New(log.NewContextLogHandler(h))

	ctx := context.Background()
	ctx = log.ContextWithAttrs(ctx, attr1)
	logger = log.LoggerWithContext(logger, ctx)

	ctx = log.ContextWithAttrs(log.ClearContextAttrs(ctx), attr2)

	logger.InfoContext(ctx, "test", "a", "b")
	assertRecordAttrs(t, h.LastRecord(), []slog.Attr{attr1, attr2, attr3})
}
