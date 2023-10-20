package featureflags

var (
	// WriteFileContents will store the contents of write observed from strace
	// data during dynamic analysis.
	WriteFileContents = new("WriteFileContents", true)

	// SaveAnalyzedPackages downloads the package archive and saves it
	// to the analyzed packages bucket (if configured) after analysis completes
	SaveAnalyzedPackages = new("SaveAnalyzedPackages", true)

	// PubSubExtender determines whether the worker uses a real GCP extender
	// for keeping messages alive during long-running processing.
	PubSubExtender = new("PubSubExtender", true)

	// CodeExecution invokes package code automatically during dynamic analysis,
	// which may uncover extra malicious behaviour. The names of executed functions,
	// methods and classes are logged to a file.
	CodeExecution = new("CodeExecution", false)

	// DisableStraceDebugLogging prevents debug logging of strace parsing during
	// dynamic analysis. Since the strace parsing produces a large amount of debug
	// output, it can be useful to disable this when the information is not needed.
	DisableStraceDebugLogging = new("DisableStraceDebugLogging", false)
)
