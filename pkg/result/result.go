package result

import (
	"encoding/json"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/strace"
	"github.com/ossf/package-analysis/pkg/api"
)

type DynamicAnalysisStraceSummary map[api.RunPhase]*StraceSummary
type DynamicAnalysisFileWrites map[api.RunPhase]*FileWrites

type DynamicAnalysisResults struct {
	StraceSummary        DynamicAnalysisStraceSummary
	FileWrites           DynamicAnalysisFileWrites
	FileWriteBufferPaths []string
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

type FileWrites []FileWriteResult

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
