package featureflags

var (
	// WriteFileContents will store the contents of write observed from strace
	// data during dynamic analysis.
	WriteFileContents = new("WriteFileContents", true)

	// SaveAnalyzedPackages downloads the package archive and saves it
	// to the analyzed packages bucket (if configured) after analysis completes
	SaveAnalyzedPackages = new("SaveAnalyzedPackages", false)

	// PubSubExtender determines whether or not the worker uses a real GCP
	// extender for keeping messages alive during long-running processing.
	PubSubExtender = new("PubSubExtender", true)

	// CodeExecution automatically invokes package code during the import phase
	// of dynamic analysis, which may uncover extra malicious behaviour.
	// A list executed function/method/class names is saved.
	CodeExecution = new("CodeExecution", true)
)
