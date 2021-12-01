package log

import (
	"bufio"
	"io"

	"go.uber.org/zap"
)

func logLine(logger *zap.SugaredLogger, level Level, l string) {
	switch level {
	case DebugLevel:
		logger.Debug(l)
	case InfoLevel:
		logger.Info(l)
	case WarnLevel:
		logger.Warn(l)
	case ErrorLevel:
		logger.Error(l)
	case DPanicLevel:
		logger.DPanic(l)
	case PanicLevel:
		logger.Panic(l)
	case FatalLevel:
		logger.Fatal(l)
	}
}

// Writer returns an io.WriteCloser that logs each line written as a single log
// entry at the given level with the supplied keysAndValues.
//
// Close() must be called to free up the resources used.
func Writer(level Level, keysAndValues ...interface{}) io.WriteCloser {
	checkInit()

	r, w := io.Pipe()
	go func() {
		err := WriteTo(level, r, keysAndValues...)
		r.Close()
		if err != nil {
			Error("Writer failed.",
				"error", err)
		}
	}()
	return w
}

// WriteTo will log each line read from r as a single log entry at the given
// level with the supplied keysAndValues.
//
// This method will block until r returns an EOF or causes an error.
func WriteTo(level Level, r io.Reader, keysAndValues ...interface{}) error {
	logger := defaultLogger.With(keysAndValues...)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		// Swallow empty lines
		if len(text) > 0 {
			logLine(logger, level, text)
		}
	}
	return scanner.Err()
}
