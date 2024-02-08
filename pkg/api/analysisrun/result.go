package analysisrun

import (
	"github.com/ossf/package-analysis/internal/analysis"
)

type (
	// DynamicAnalysisStraceSummary holds system calls made during each analysis phase,
	// obtained by strace monitoring.
	DynamicAnalysisStraceSummary map[DynamicPhase]*StraceSummary

	// DynamicAnalysisFileWritesSummary holds a summary of files written by all processes
	// under analysis, during each analysis phase. This includes a list of paths written to,
	// and counts of bytes written each time. Write data is obtained via strace monitoring.
	DynamicAnalysisFileWritesSummary map[DynamicPhase]*FileWritesSummary

	// DynamicAnalysisFileWriteBufferIds holds IDs (names) for each recorded write operation
	// during each analysis phase. These names correspond to files in a zip archive that contain
	// the actual write buffer contents.
	DynamicAnalysisFileWriteBufferIds map[DynamicPhase][]string

	// DynamicAnalysisExecutionLog contains a record of which package symbols (e.g. modules,
	// functions, classes) were discovered during the 'execute' analysis phase, and the results
	// of attempts to call or instantiate them.
	DynamicAnalysisExecutionLog string
)

// DynamicAnalysisRecord is a generic top-level struct which is used to produce JSON results
// files for dynamic analysis in the current schema format. This format is used for
// strace data, file write summary data and execution log data.
type DynamicAnalysisRecord struct {
	Package          Key   `json:"Package"`
	CreatedTimestamp int64 `json:"CreatedTimestamp"`
	Analysis         any   `json:"Analysis"`
}

// DynamicAnalysisStraceRecord is a specialisation of DynamicAnalysisRecord that can be used for
// deserializing JSON files from the original strace-only dynamic analysis results.
type DynamicAnalysisStraceRecord struct {
	Package          Key                          `json:"Package"`
	CreatedTimestamp int64                        `json:"CreatedTimestamp"`
	Analysis         DynamicAnalysisStraceSummary `json:"Analysis"`
}

// DynamicAnalysisData holds all data obtained from running dynamic analysis.
type DynamicAnalysisData struct {
	StraceSummary      DynamicAnalysisStraceSummary
	FileWritesSummary  DynamicAnalysisFileWritesSummary
	FileWriteBufferIds DynamicAnalysisFileWriteBufferIds
	ExecutionLog       DynamicAnalysisExecutionLog
}

type StraceSummary struct {
	Status   analysis.Status
	Stdout   []byte
	Stderr   []byte
	Files    []FileResult
	Sockets  []SocketResult
	Commands []CommandResult
	DNS      []DNSResult
}

type FileWritesSummary []FileWriteResult

type FileWriteResult struct {
	Path      string
	WriteInfo []WriteInfo
}

type WriteInfo struct {
	WriteBufferId string
	BytesWritten  int64
}

type FileResult struct {
	Path   string
	Read   bool
	Write  bool
	Delete bool
}

type SocketResult struct {
	Address   string
	Port      int
	Hostnames []string
}

type CommandResult struct {
	Command     []string
	Environment []string
}

type DNSQueries struct {
	Hostname string
	Types    []string
}

type DNSResult struct {
	Class   string
	Queries []DNSQueries
}
