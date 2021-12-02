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
	"strings"

	"github.com/blendle/zapdriver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level represents a specific logging level. It wraps zapcore.Level.
type Level zapcore.Level

const (
	DebugLevel  Level = Level(zapcore.DebugLevel)
	InfoLevel   Level = Level(zapcore.InfoLevel)
	WarnLevel   Level = Level(zapcore.WarnLevel)
	ErrorLevel  Level = Level(zapcore.ErrorLevel)
	DPanicLevel Level = Level(zapcore.DPanicLevel)
	PanicLevel  Level = Level(zapcore.PanicLevel)
	FatalLevel  Level = Level(zapcore.FatalLevel)
)

// LoggingEnv is used to represent a specific configuration used by a given
// environment.
type LoggingEnv string

// String implements the Stringer interface.
func (e LoggingEnv) String() string {
	return string(e)
}

const (
	LoggingEnvDev  LoggingEnv = "dev"
	LoggingEnvProd LoggingEnv = "prod"
)

var (
	defaultLogger     *zap.SugaredLogger
	defaultLoggingEnv LoggingEnv = LoggingEnvDev
)

// Initalize the logger for logging.
//
// Passing in "true" will use Zap's default production configuration, while
// "false" will use the default development configuration.
//
// Note: this method MUST be called before any other method in this package.
func Initalize(env string) {
	var err error
	var logger *zap.Logger
	switch strings.ToLower(env) {
	case LoggingEnvProd.String():
		defaultLoggingEnv = LoggingEnvProd
		config := zapdriver.NewProductionConfig()
		// Make sure sampling is disabled.
		config.Sampling = nil
		// Build the logger and ensure we use the zapdriver Core so that labels
		// are handled correctly.
		logger, err = config.Build(zapdriver.WrapCore())
	case LoggingEnvDev.String():
		fallthrough
	default:
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		golog.Panic(err)
	}
	zap.RedirectStdLog(logger)
	defaultLogger = logger.WithOptions(zap.AddCallerSkip(1)).Sugar()
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

// Label is a convenience wrapper for zapdriver.Label if the LoggingEnv used
// is LoggingEnvProd. Otherwise it will wrap zap.String.
func Label(key, value string) zap.Field {
	if defaultLoggingEnv == LoggingEnvProd {
		return zapdriver.Label(key, value)
	} else {
		return zap.String(key, value)
	}
}
