package featureflags

var (
	// WriteFileContents will store the contents of write observed from strace
	// data during dynamic analysis.
	WriteFileContents = new("WriteFileContents", true)

	// SaveAnalyzedPackages will save analyzed package.
	SaveAnalyzedPackages = new("SaveAnalyzedPackages", false)

	// PubSubExtender determines whether or not the worker uses a real GCP
	// extender for keeping messages alive during long-running processing.
	PubSubExtender = new("PubSubExtender", true)
)
