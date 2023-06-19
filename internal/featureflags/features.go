package featureflags

var (
	// WriteFileContents will store the contents of write observed from strace
	// data during dynamic analysis.
	WriteFileContents = new("WriteFileContents", true)
	// SaveAnalyzedPackages will save analyzed package.
	SaveAnalyzedPackages = new("SaveAnalyzedPackages", false)
)
