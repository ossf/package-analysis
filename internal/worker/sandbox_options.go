package worker

import (
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
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
func DynamicSandboxOptions(ecosystem pkgecosystem.Ecosystem) []sandbox.Option {
	return []sandbox.Option{
		sandbox.Image(defaultDynamicAnalysisImage),
		sandbox.Command(defaultDynamicAnalysisCommand[ecosystem]),
		sandbox.EnableStrace(),
		sandbox.EnableRawSockets(),
		sandbox.EnablePacketLogging(),
		sandbox.LogStdOut(),
		sandbox.LogStdErr(),
	}
}
