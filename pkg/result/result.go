package result

import (
	"encoding/json"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/strace"
	"github.com/ossf/package-analysis/pkg/api"
)

type DynamicAnalysisStraceSummary map[api.RunPhase]*StraceSummary
type DynamicAnalysisFileWritesSummary map[api.RunPhase]*FileWritesSummary
type DynamicAnalysisFileWriteBufferPaths map[api.RunPhase][]string

type DynamicAnalysisResults struct {
	StraceSummary     DynamicAnalysisStraceSummary
	FileWritesSummary DynamicAnalysisFileWritesSummary
	// Paths to files on disk that contain write buffer data from write syscalls
	FileWriteBufferPaths DynamicAnalysisFileWriteBufferPaths
}

type StaticAnalysisResults = json.RawMessage

type StraceSummary struct {
	Status   analysis.Status
	Stdout   []byte
	Stderr   []byte
	Files    []FileResult
	Sockets  []SocketResult
	Commands []CommandResult
	DNS      []DnsResult
}

type FileWritesSummary []FileWriteResult

type FileWriteResult struct {
	Path      string
	WriteInfo strace.WriteInfo
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

type DnsQueries struct {
	Hostname string
	Types    []string
}

type DnsResult struct {
	Class   string
	Queries []DnsQueries
}
