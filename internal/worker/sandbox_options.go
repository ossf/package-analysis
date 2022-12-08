package worker

import (
	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/sandbox"
)

// DefaultSandboxOptions initialises sandbox options necessary to run the given analysis mode
func DefaultSandboxOptions(mode analysis.Mode, imageTag string) []sandbox.Option {
	switch mode {
	case analysis.Dynamic:
		return []sandbox.Option{
			sandbox.Tag(imageTag),
			sandbox.EnableStrace(),
			sandbox.EnableRawSockets(),
			sandbox.EnablePacketLogging(),
			sandbox.LogStdOut(),
			sandbox.LogStdErr(),
		}
	case analysis.Static:
		return []sandbox.Option{
			sandbox.Tag(imageTag),
			sandbox.EchoStdErr(),
		}
	default:
		return []sandbox.Option{}
	}
}
