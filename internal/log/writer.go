package log

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"unicode"
)

// NewWriter returns an io.WriteCloser that logs each line written as a single
// log entry at the given level with the supplied keysAndValues.
//
// Close() must be called to free up the resources used and flush any unwritten
// log entries to the logger.
func NewWriter(ctx context.Context, logger *slog.Logger, level slog.Level) io.WriteCloser {
	return &writer{
		ctx:    ctx,
		logger: logger,
		level:  level,
	}
}

type writer struct {
	ctx    context.Context
	logger *slog.Logger
	level  slog.Level
	buffer bytes.Buffer
}

// Write implements the io.Writer interface.
//
// Each line of bytes written appears as a log entry.
func (w *writer) Write(p []byte) (int, error) {
	written := 0
	for {
		if len(p) == 0 {
			// p is now empty, so exit with the bytes written
			return written, nil
		}
		i := bytes.IndexByte(p, '\n')
		if i == -1 {
			// No more newlines to consume, so save the buffer and return
			n, err := w.buffer.Write(p)
			return written + n, err
		}
		// Append to the buffer.
		n, err := w.buffer.Write(p[:i])
		written += n
		if err != nil {
			return written, err
		}
		// Update the input and consume the newline
		p = p[i+1:]
		written += 1
		// Dump the buffer to the log
		line := w.buffer.Bytes()
		// Trim any trailing space - this won't include the newline
		line = bytes.TrimRightFunc(line, unicode.IsSpace)
		// Swallow any empty lines
		if len(line) > 0 {
			w.logger.Log(w.ctx, w.level, string(line))
		}
		// Reset the buffer.
		w.buffer.Reset()
	}
}

// Close implements the io.Closer interface.
//
// Any unwritten bytes written as a final log entry.
func (w *writer) Close() error {
	if w.buffer.Len() > 0 {
		w.logger.Log(w.ctx, w.level, w.buffer.String())
		w.buffer.Reset()
	}
	return nil
}
