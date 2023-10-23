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

	// StraceDebugLogDir is a hardcoded directory that can be used to store
	// the strace debug log, if the strace debug logging feature is enabled
	StraceDebugLogDir = "/straceLogs"
)

var (
	defaultLoggingEnv LoggingEnv = LoggingEnvDev
)

func DefaultLoggingEnv() LoggingEnv {
	return defaultLoggingEnv
}

// Initialize the logger for logging.
//
// Passing in "true" will use Zap's default production configuration, while
// "false" will use the default development configuration.
//
// Note: this method MUST be called before any other method in this package.
func Initialize(env string) {
	// TODO: replace zap entirely with native slog.
	// Note that zap currently provides some useful features, such as prod and
	// dev environments, standard logger replacement, and GCP StackDriver
	// integration. Since log/slog is so new, many of the same capabilities are
	// yet to receive good support in third-party libraries.
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
}

// Label causes attributes written by zapdriver to be marked as labels inside
// StackDriver when LoggingEnv is LoggingEnvProd. Otherwise it wraps slog.String.
func Label(key, value string) slog.Attr {
	if defaultLoggingEnv == LoggingEnvProd {
		return slog.String("labels."+key, value)
	} else {
		return slog.String(key, value)
	}
}
