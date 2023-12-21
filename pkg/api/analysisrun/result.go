package analysisrun

import (
	"github.com/ossf/package-analysis/internal/analysis"
)

type (
	DynamicAnalysisStraceSummary      map[DynamicPhase]*StraceSummary
	DynamicAnalysisFileWritesSummary  map[DynamicPhase]*FileWritesSummary
	DynamicAnalysisFileWriteBufferIds map[DynamicPhase][]string
	DynamicAnalysisExecutionLog       string
)

// DynamicAnalysisRecord is the top-level struct which is serialised to produce JSON results files
// for dynamic analysis.
type DynamicAnalysisRecord struct {
	Package          Key   `json:"Package"`
	CreatedTimestamp int64 `json:"CreatedTimestamp"`
	Analysis         any   `json:"Analysis"`
}

type DynamicAnalysisResults struct {
	StraceSummary     DynamicAnalysisStraceSummary
	FileWritesSummary DynamicAnalysisFileWritesSummary
	// Ids that correlate to the name of the file that saves the actual write buffer contents.
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
