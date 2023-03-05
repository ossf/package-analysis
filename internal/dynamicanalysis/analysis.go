package dynamicanalysis

import (
	"fmt"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/dnsanalyzer"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/packetcapture"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/strace"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
)

const (
	maxOutputLines = 20
	maxOutputBytes = 4 * 1024
)

type Result struct {
	StraceSummary analysisrun.StraceSummary
	FileWrites    analysisrun.FileWritesSummary
	// Paths to files on disk that contain write buffer data from write syscalls
	FileWriteBufferPaths []string
}

var resultError = &Result{
	StraceSummary: analysisrun.StraceSummary{
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

	analysisResult := Result{
		StraceSummary: analysisrun.StraceSummary{
			Status: analysis.StatusForRunResult(r),
			Stdout: lastLines(r.Stdout(), maxOutputLines, maxOutputBytes),
			Stderr: lastLines(r.Stderr(), maxOutputLines, maxOutputBytes),
		},
	}
	analysisResult.setData(straceResult, dns)
	return &analysisResult, nil
}

func (d *Result) setData(straceResult *strace.Result, dns *dnsanalyzer.DNSAnalyzer) {
	for _, f := range straceResult.Files() {
		d.StraceSummary.Files = append(d.StraceSummary.Files, analysisrun.FileResult{
			Path:   f.Path,
			Read:   f.Read,
			Write:  f.Write,
			Delete: f.Delete,
		})
		if len(f.WriteInfo) > 0 {
			w := analysisrun.FileWritesSummary{Path: f.Path}
			for _, wi := range f.WriteInfo {
				w.WriteInfo = append(w.WriteInfo, analysisrun.WriteInfo{
					WriteBufferId: wi.WriteBufferId,
					BytesWritten:  wi.BytesWritten,
				})
			}
			d.FileWritesSummary = append(d.FileWritesSummary, w)
			for _, writeBufferPath := range f.WriteBufferPaths {
				d.FileWriteBufferPaths = append(d.FileWriteBufferPaths, writeBufferPath)
			}
		}
	}

	for _, s := range straceResult.Sockets() {
		d.StraceSummary.Sockets = append(d.StraceSummary.Sockets, analysisrun.SocketResult{
			Address:   s.Address,
			Port:      s.Port,
			Hostnames: dns.Hostnames(s.Address),
		})
	}

	for _, c := range straceResult.Commands() {
		d.StraceSummary.Commands = append(d.StraceSummary.Commands, analysisrun.CommandResult{
			Command:     c.Command,
			Environment: c.Env,
		})
	}

	for dnsClass, queries := range dns.Questions() {
		c := analysisrun.DNSResult{Class: dnsClass}
		for host, types := range queries {
			c.Queries = append(c.Queries, analysisrun.DNSQueries{
				Hostname: host,
				Types:    types,
			})
		}
		d.StraceSummary.DNS = append(d.StraceSummary.DNS, c)
	}
}
