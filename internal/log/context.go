package log

import (
	"context"
	"log/slog"
)

var contextAttrSliceKey = struct{}{}

func attrSliceFromContext(ctx context.Context) []slog.Attr {
	if v := ctx.Value(contextAttrSliceKey); v != nil {
		return v.([]slog.Attr)
	}
	return nil
}

// ContextWithAttrs is used to add attrs to the context so they are included
// when logs are output.
func ContextWithAttrs(ctx context.Context, attr ...slog.Attr) context.Context {
	if len(attr) == 0 {
		return ctx
	}
	attrSlice := attrSliceFromContext(ctx)
	attrSlice = append(attrSlice, attr...)
	return context.WithValue(ctx, contextAttrSliceKey, attrSlice)
}

type contextLogHandler struct {
	slog.Handler
}

func (h *contextLogHandler) Handle(ctx context.Context, r slog.Record) error {
	attrSlice := attrSliceFromContext(ctx)
	if attrSlice != nil {
		r.AddAttrs(attrSlice...)
	}
	return h.Handler.Handle(ctx, r)
}

// NewContextLogHandler returns a new slog.Handler that will pass the attrs set
// using ContextWithAttrs to handler when Handle is called.
func NewContextLogHandler(handler slog.Handler) slog.Handler {
	return &contextLogHandler{
		Handler: handler,
	}
}
