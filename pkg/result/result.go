package result

import (
	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/strace"
)

type DynamicAnalysisStraceSummary map[pkgecosystem.RunPhase]*StraceSummary
type DynamicAnalysisFileWrites map[pkgecosystem.RunPhase]*FileWrites

type DynamicAnalysisResults struct {
	StraceSummary DynamicAnalysisStraceSummary
	FileWrites    DynamicAnalysisFileWrites
}

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