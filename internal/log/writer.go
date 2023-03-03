package log

import (
	"bytes"
	"io"
	"unicode"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Writer returns an io.WriteCloser that logs each line written as a single
// log entry at the given level with the supplied keysAndValues.
//
// Close() must be called to free up the resources used and flush any unwritten
// log entries to the logger.
//
// Deprecated: use NewWriter with zap.Logger and zapcore.Level directly.
func Writer(level Level, keysAndValues ...interface{}) io.WriteCloser {
	checkInit()
	logger := defaultLogger.With(keysAndValues...).Desugar()
	return NewWriter(logger, zapcore.Level(level))
}

// NewWriter returns an io.WriteCloser that logs each line written as a single
// log entry at the given level with the supplied keysAndValues.
//
// Close() must be called to free up the resources used and flush any unwritten
// log entries to the logger.
func NewWriter(logger *zap.Logger, level zapcore.Level) io.WriteCloser {
	return &writer{
		logger: logger,
		level:  level,
	}
}

type writer struct {
	logger *zap.Logger
	level  zapcore.Level
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
			w.logger.Log(w.level, string(line))
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
		w.logger.Log(w.level, w.buffer.String())
		w.buffer.Reset()
	}
	return nil
}
