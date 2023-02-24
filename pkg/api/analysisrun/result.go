package analysisrun

import (
	"encoding/json"

	"github.com/ossf/package-analysis/internal/analysis"
)

type (
	DynamicAnalysisStraceSummary map[DynamicPhase]*StraceSummary
	DynamicAnalysisFileWrites    map[DynamicPhase]*FileWrites
)

type DynamicAnalysisResults struct {
	StraceSummary DynamicAnalysisStraceSummary
	FileWrites    DynamicAnalysisFileWrites
}

type StaticAnalysisResults = json.RawMessage

type StraceSummary struct {
	Status   analysis.Status
	Stdout   []byte
	Stderr   []byte
	Files    []FileResult
	Sockets  []SocketResult
	Commands []CommandResult
	DNS      []DNSResult
}

type FileWrites []FileWriteResult

type FileWriteResult struct {
	Path      string
	WriteInfo []WriteInfo
}

type WriteInfo struct {
	BytesWritten int64
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
