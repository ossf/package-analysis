package dynamicanalysis

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/dnsanalyzer"
	"github.com/ossf/package-analysis/internal/packetcapture"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/strace"
	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
)

const (
	maxOutputBytes = 4 * 1024
)

type Result struct {
	StraceSummary     analysisrun.StraceSummary
	FileWritesSummary analysisrun.FileWritesSummary
	// IDs that correlate to the name of the file that saves the actual write buffer contents.
	// We save this separately so that we don't need to dig through the FileWritesSummary later on.
	FileWriteBufferIds []string
}

var resultError = &Result{
	StraceSummary: analysisrun.StraceSummary{
		Status: analysis.StatusErrorOther,
	},
}

func Run(ctx context.Context, sb sandbox.Sandbox, command string, args []string, straceLogger *slog.Logger) (*Result, error) {
	slog.InfoContext(ctx, "Running dynamic analysis", "args", args)

	slog.DebugContext(ctx, "Preparing packet capture")
	pcap := packetcapture.New(sandbox.NetworkInterface)

	dns := dnsanalyzer.New()
	pcap.RegisterReceiver(dns)
	if err := pcap.Start(); err != nil {
		return resultError, fmt.Errorf("failed to start packet capture (%w)", err)
	}
	defer pcap.Close()

	// Run the command
	slog.DebugContext(ctx, "Running dynamic analysis command",
		"command", command,
		"args", args)
	r, err := sb.Run(ctx, command, args...)
	if err != nil {
		return resultError, fmt.Errorf("sandbox failed (%w)", err)
	}

	slog.DebugContext(ctx, "Stop the packet capture")
	pcap.Close()

	// Grab the log file
	slog.DebugContext(ctx, "Parsing the strace log")
	l, err := r.Log()
	if err != nil {
		return resultError, fmt.Errorf("failed to open strace log (%w)", err)
	}
	defer l.Close()

	straceResult, err := strace.Parse(ctx, l, straceLogger)
	if err != nil {
		return resultError, fmt.Errorf("strace parsing failed (%w)", err)
	}

	analysisResult := Result{
		StraceSummary: analysisrun.StraceSummary{
			Status: analysis.StatusForRunResult(r),
			Stdout: utils.LastNBytes(r.Stdout(), maxOutputBytes),
			Stderr: utils.LastNBytes(r.Stderr(), maxOutputBytes),
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
			w := analysisrun.FileWriteResult{Path: f.Path}
			for _, wi := range f.WriteInfo {
				w.WriteInfo = append(w.WriteInfo, analysisrun.WriteInfo{
					WriteBufferId: wi.WriteBufferId,
					BytesWritten:  wi.BytesWritten,
				})
				d.FileWriteBufferIds = append(d.FileWriteBufferIds, wi.WriteBufferId)
			}
			d.FileWritesSummary = append(d.FileWritesSummary, w)
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
