package featureflags

var (
	// WriteFileContents will store the contents of write observed from strace
	// data during dynamic analysis.
	WriteFileContents = new("WriteFileContents", true)
	// AnalyzedPackage will save analyzed package.
	AnalyzedPackage = new("AnalyzedPackage", false)
)
