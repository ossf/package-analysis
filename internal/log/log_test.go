package log_test

import (
	"context"
	"log/slog"
)

type testHandler struct {
	slog.Handler

	root    *testHandler
	records []slog.Record
	attrs   []slog.Attr
}

func (h *testHandler) getRoot() *testHandler {
	if h.root == nil {
		return h
	}
	return h.root
}

func (h *testHandler) LastRecord() slog.Record {
	root := h.getRoot()
	l := len(root.records)
	if l == 0 {
		return slog.Record{}
	}
	return root.records[l-1]
}

func (h *testHandler) All() []slog.Record {
	root := h.getRoot()
	return root.records
}

func (h *testHandler) Len() int {
	root := h.getRoot()
	return len(root.records)
}

func (h *testHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *testHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(h.attrs...)
	root := h.getRoot()
	root.records = append(h.getRoot().records, r)
	return nil
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &testHandler{
		root:  h.getRoot(),
		attrs: append(h.attrs, attrs...),
	}
}
