package worker

import (
	"github.com/ossf/package-analysis/internal/sandbox"
)

// StaticSandboxOptions provides a set of sandbox options necessary to run the
// static analysis sandboxes.
func StaticSandboxOptions() []sandbox.Option {
	return []sandbox.Option{
		sandbox.Image(defaultStaticAnalysisImage),
		sandbox.EchoStdErr(),
	}
}

// DynamicSandboxOptions provides a set of sandbox options necessary to run
// dynamic analysis sandboxes.
func DynamicSandboxOptions() []sandbox.Option {
	return []sandbox.Option{
		sandbox.Image(defaultDynamicAnalysisImage),
		sandbox.EnableStrace(),
		sandbox.EnableRawSockets(),
		sandbox.EnablePacketLogging(),
		sandbox.LogStdOut(),
		sandbox.LogStdErr(),
	}
}
