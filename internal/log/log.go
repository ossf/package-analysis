// Package log wraps Uber's Zap logging library to make it easy to use across
// the project.
//
// Initialize() MUST be called before the first logging statement, if it is not
// called the command will panic and exit.
//
// See the Zap docs for more details: https://pkg.go.dev/go.uber.org/zap
package log

import (
	golog "log"

	"go.uber.org/zap"
)

var (
	defaultLogger *zap.SugaredLogger
)

// Initalize the logger for logging.
//
// Passing in "true" will use Zap's default production configuration, while
// "false" will use the default development configuration.
//
// Note: this method MUST be called before any other method in this package.
func Initalize(prod bool) {
	var err error
	var logger *zap.Logger
	if prod {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		golog.Panic(err)
	}
	defaultLogger = logger.Sugar()
}

func checkInit() {
	if defaultLogger == nil {
		golog.Panic("Must call log.Initialize(...) before logging.")
	}
}

// Debug is a convenience wrapper for calling Debugw on the default
// zap.SugaredLogger instance
func Debug(msg string, keysAndValues ...interface{}) {
	checkInit()
	defaultLogger.Debugw(msg, keysAndValues...)
}

// Info is a convenience wrapper for calling Infow on the default
// zap.SugaredLogger instance
func Info(msg string, keysAndValues ...interface{}) {
	checkInit()
	defaultLogger.Infow(msg, keysAndValues...)
}

// Warn is a convenience wrapper for calling Warnw on the default
// zap.SugaredLogger instance
func Warn(msg string, keysAndValues ...interface{}) {
	checkInit()
	defaultLogger.Warnw(msg, keysAndValues...)
}

// Error is a convenience wrapper for calling Errorw on the default
// zap.SugaredLogger instance
func Error(msg string, keysAndValues ...interface{}) {
	checkInit()
	defaultLogger.Errorw(msg, keysAndValues...)
}

// Fatal is a convenience wrapper for calling Fatalw on the default
// zap.SugaredLogger instance
func Fatal(msg string, keysAndValues ...interface{}) {
	checkInit()
	defaultLogger.Fatalw(msg, keysAndValues...)
}

// Panic is a convenience wrapper for calling Panicw on the default
// zap.SugaredLogger instance
func Panic(msg string, keysAndValues ...interface{}) {
	checkInit()
	defaultLogger.Panicw(msg, keysAndValues...)
}

// DPanic is a convenience wrapper for calling DPanicw on the default
// zap.SugaredLogger instance
func DPanic(msg string, keysAndValues ...interface{}) {
	checkInit()
	defaultLogger.DPanicw(msg, keysAndValues...)
}
