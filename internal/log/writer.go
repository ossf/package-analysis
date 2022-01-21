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
func WriteTo(level Level, rd io.Reader, keysAndValues ...interface{}) error {
	logger := defaultLogger.With(keysAndValues...)
	r := bufio.NewReader(rd)
	for {
		line, err := r.ReadBytes('\n')
		l := len(line)
		// Trim any trailing \n or \r\n from the buffer.
		if l > 0 && line[l-1] == '\n' {
			l = l - 1
			if l > 0 && line[l-1] == '\r' {
				l = l - 1
			}
		}
		// Swallow empty lines
		if l > 0 {
			logLine(logger, level, string(line[:l]))
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
