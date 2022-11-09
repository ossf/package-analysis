package dynamicanalysis

import (
	"fmt"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/dnsanalyzer"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/packetcapture"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/strace"
)

const (
	maxOutputLines = 20
	maxOutputBytes = 4 * 1024
)

type fileResult struct {
	Path   string
	Read   bool
	Write  bool
	Delete bool
}

type FileWrites struct {
	Path      string
	WriteInfo []strace.WriteInfo
}

type socketResult struct {
	Address   string
	Port      int
	Hostnames []string
}

type commandResult struct {
	Command     []string
	Environment []string
}

type dnsQueries struct {
	Hostname string
	Types    []string
}

type dnsResult struct {
	Class   string
	Queries []dnsQueries
}

type StraceSummary struct {
	Status   analysis.Status
	Stdout   []byte
	Stderr   []byte
	Files    []fileResult
	Sockets  []socketResult
	Commands []commandResult
	DNS      []dnsResult
}

type Result struct {
	StraceSummary StraceSummary
	FileWrites    []FileWrites
}

var resultError = &Result{
	StraceSummary: StraceSummary{
		Status: analysis.StatusErrorOther,
	},
}

func Run(sb sandbox.Sandbox, args []string) (*Result, error) {
	log.Info("Running dynamic analysis",
		"args", args)

	log.Debug("Preparing packet capture")
	pcap := packetcapture.New(sandbox.NetworkInterface)

	dns := dnsanalyzer.New()
	pcap.RegisterReceiver(dns)
	if err := pcap.Start(); err != nil {
		return resultError, fmt.Errorf("failed to start packet capture (%w)", err)
	}
	defer pcap.Close()

	// Run the command
	log.Debug("Running dynamic analysis command",
		"args", args)
	r, err := sb.Run(args...)
	if err != nil {
		return resultError, fmt.Errorf("sandbox failed (%w)", err)
	}

	log.Debug("Stop the packet capture")
	pcap.Close()

	// Grab the log file
	log.Debug("Parsing the strace log")
	l, err := r.Log()
	if err != nil {
		return resultError, fmt.Errorf("failed to open strace log (%w)", err)
	}
	defer l.Close()

	straceResult, err := strace.Parse(l)
	if err != nil {
		return resultError, fmt.Errorf("strace parsing failed (%w)", err)
	}

	result := Result{
		StraceSummary: StraceSummary{
			Status: analysis.StatusForRunResult(r),
			Stdout: lastLines(r.Stdout(), maxOutputLines, maxOutputBytes),
			Stderr: lastLines(r.Stderr(), maxOutputLines, maxOutputBytes),
		},
	}
	result.setData(straceResult, dns)
	return &result, nil
}

func (d *Result) setData(straceResult *strace.Result, dns *dnsanalyzer.DNSAnalyzer) {
	for _, f := range straceResult.Files() {
		d.StraceSummary.Files = append(d.StraceSummary.Files, fileResult{
			Path:   f.Path,
			Read:   f.Read,
			Write:  f.Write,
			Delete: f.Delete,
		})
		if len(f.WriteInfo) > 0 {
			d.FileWrites = append(d.FileWrites, FileWrites{f.Path, f.WriteInfo})
		}
	}

	for _, s := range straceResult.Sockets() {
		d.StraceSummary.Sockets = append(d.StraceSummary.Sockets, socketResult{
			Address:   s.Address,
			Port:      s.Port,
			Hostnames: dns.Hostnames(s.Address),
		})
	}

	for _, c := range straceResult.Commands() {
		d.StraceSummary.Commands = append(d.StraceSummary.Commands, commandResult{
			Command:     c.Command,
			Environment: c.Env,
		})
	}

	for dnsClass, queries := range dns.Questions() {
		c := dnsResult{Class: dnsClass}
		for host, types := range queries {
			c.Queries = append(c.Queries, dnsQueries{
				Hostname: host,
				Types:    types,
			})
		}
		d.StraceSummary.DNS = append(d.StraceSummary.DNS, c)
	}
}
