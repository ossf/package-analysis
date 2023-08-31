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
	"log/slog"
	"strings"

	"github.com/blendle/zapdriver"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

// Level represents a specific logging level. It wraps zapcore.Level.
//
// Deprecated: use zapcore.Level directly.
type Level zapcore.Level

const (
	DebugLevel Level = Level(zapcore.DebugLevel)
	InfoLevel  Level = Level(zapcore.InfoLevel)
	WarnLevel  Level = Level(zapcore.WarnLevel)
	ErrorLevel Level = Level(zapcore.ErrorLevel)
	FatalLevel Level = Level(zapcore.FatalLevel)
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
	// Default logger is the legacy global logger for package analysis
	// Deprecated: do not use global logger
	defaultLogger     *zap.SugaredLogger
	defaultLoggingEnv LoggingEnv = LoggingEnvDev
)

// Deprecated: do not use global logger
func GetDefaultLogger() *zap.SugaredLogger {
	return defaultLogger
}

// Initialize the logger for logging.
//
// Passing in "true" will use Zap's default production configuration, while
// "false" will use the default development configuration.
//
// Note: this method MUST be called before any other method in this package.
func Initialize(env string) *zap.SugaredLogger {
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
	// Ensure slog.Default logs to the same destination as zap.
	slogger := slog.New(NewContextLogHandler(zapslog.NewHandler(logger.Core(), &zapslog.HandlerOptions{
		AddSource: true,
	})))
	slog.SetDefault(slogger)

	logger = logger.WithOptions(zap.AddCallerSkip(1))
	defaultLogger = logger.Sugar() // Set defaultLogger to provide legacy support
	return defaultLogger
}

func checkInit() {
	if defaultLogger == nil {
		golog.Panic("Must call log.Initialize(...) before logging.")
	}
}

// Debug is a convenience wrapper for calling Debugw on the default
// zap.SugaredLogger instance.
//
// Deprecated: Call slog.DebugContext instead.
func Debug(msg string, keysAndValues ...interface{}) {
	checkInit()
	defaultLogger.Debugw(msg, keysAndValues...)
}

// Info is a convenience wrapper for calling Infow on the default
// zap.SugaredLogger instance.
//
// Deprecated: Call slog.InfoContext instead.
func Info(msg string, keysAndValues ...interface{}) {
	checkInit()
	defaultLogger.Infow(msg, keysAndValues...)
}

// Warn is a convenience wrapper for calling Warnw on the default
// zap.SugaredLogger instance.
//
// Deprecated: Call slog.WarnContext instead.
func Warn(msg string, keysAndValues ...interface{}) {
	checkInit()
	defaultLogger.Warnw(msg, keysAndValues...)
}

// Error is a convenience wrapper for calling Errorw on the default
// zap.SugaredLogger instance.
//
// Deprecated: Call slog.ErrorContext instead.
func Error(msg string, keysAndValues ...interface{}) {
	checkInit()
	defaultLogger.Errorw(msg, keysAndValues...)
}

// Fatal is a convenience wrapper for calling Fatalw on the default
// zap.SugaredLogger instance.
//
// Deprecated: Call slog.ErrorContext, followed by os.Exit instead.
func Fatal(msg string, keysAndValues ...interface{}) {
	checkInit()
	defaultLogger.Fatalw(msg, keysAndValues...)
}

// Label is a convenience wrapper for zapdriver.Label if the LoggingEnv used
// is LoggingEnvProd. Otherwise it will wrap zap.String.
//
// Deprecated: Call LabelAttr instead.
func Label(key, value string) zap.Field {
	if defaultLoggingEnv == LoggingEnvProd {
		return zapdriver.Label(key, value)
	} else {
		return zap.String(key, value)
	}
}

// LabelAttr causes attributes written by zapdriver to be marked as labels inside
// StackDriver when LoggingEnv is LoggingEnvProd. Otherwise it wraps slog.String.
func LabelAttr(key, value string) slog.Attr {
	if defaultLoggingEnv == LoggingEnvProd {
		return slog.String("labels."+key, value)
	} else {
		return slog.String(key, value)
	}
}
