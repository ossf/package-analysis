package log

import (
	"context"
	"log/slog"
)

type attrSliceContextKey struct{}

func attrSliceFromContext(ctx context.Context) []slog.Attr {
	if v := ctx.Value(attrSliceContextKey{}); v != nil {
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
	attrSlice := append(attrSliceFromContext(ctx), attr...)
	return context.WithValue(ctx, attrSliceContextKey{}, attrSlice)
}

func ClearContextAttrs(ctx context.Context) context.Context {
	attrSlice := attrSliceFromContext(ctx)
	if attrSlice == nil {
		return ctx
	}
	return context.WithValue(ctx, attrSliceContextKey{}, nil)
}

// LoggerWithContext returns a logger with any attrs in the context passed to
// the logger.
//
// Note: duplicate attributes may be logged if ctx, or a descendent, is used
// later in a call to (Debug|Info|Warn|Error)Context on the returned slog.Logger.
//
// If the same context is needed, call ClearContextAttrs on the context to avoid
// logging the attrs again.
func LoggerWithContext(logger *slog.Logger, ctx context.Context) *slog.Logger {
	attrSlice := attrSliceFromContext(ctx)
	if len(attrSlice) == 0 {
		return logger
	}
	return slog.New(logger.Handler().WithAttrs(attrSlice))
}

type contextLogHandler struct {
	handler slog.Handler
}

func (h *contextLogHandler) Handle(ctx context.Context, r slog.Record) error {
	attrSlice := attrSliceFromContext(ctx)
	if len(attrSlice) > 0 {
		r.AddAttrs(attrSlice...)
	}
	return h.handler.Handle(ctx, r)
}

func (h *contextLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextLogHandler{
		handler: h.handler.WithAttrs(attrs),
	}
}

func (h *contextLogHandler) WithGroup(name string) slog.Handler {
	return &contextLogHandler{
		handler: h.handler.WithGroup(name),
	}
}

func (h *contextLogHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.handler.Enabled(ctx, l)
}

func NewContextLogHandler(handler slog.Handler) slog.Handler {
	return &contextLogHandler{
		handler: handler,
	}
}
