package analysis

import (
	"encoding/json"
	"fmt"

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

type status string

// MarshalJSON implements the json.Marshaler interface.
func (s status) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

const (
	// StatusCompleted indicates that the analysis run completed successfully.
	StatusCompleted = status("completed")

	// StatusErrorTimeout indicates that the analysis was aborted due to a
	// timeout.
	StatusErrorTimeout = status("error_timeout")

	// StatusErrorAnalysis indicates that the package being analyzed failed
	// while running the specified command.
	//
	// The Stdout and Stderr in the Result should be consulted to understand
	// further why it failed.
	StatusErrorAnalysis = status("error_analysis")

	// StatusErrorOther indicates an error during some part of the analysis
	// excluding errors covered by other statuses.
	StatusErrorOther = status("error_other")
)

func statusForRunResult(r *sandbox.RunResult) status {
	switch r.Status() {
	case sandbox.RunStatusSuccess:
		return StatusCompleted
	case sandbox.RunStatusFailure:
		return StatusErrorAnalysis
	case sandbox.RunStatusTimeout:
		return StatusErrorTimeout
	default:
		return StatusErrorOther
	}
}

type fileResult struct {
	Path  string
	Read  bool
	Write bool
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

type Result struct {
	Status   status
	Stdout   []byte
	Stderr   []byte
	Files    []fileResult
	Sockets  []socketResult
	Commands []commandResult
	DNS      []dnsResult
}

var (
	resultError = &Result{Status: StatusErrorOther}
)

func Run(sb sandbox.Sandbox, args []string) (*Result, error) {
	log.Info("Running analysis",
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
	log.Debug("Running the command",
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
		Status: statusForRunResult(r),
		Stdout: lastLines(r.Stdout(), maxOutputLines, maxOutputBytes),
		Stderr: lastLines(r.Stderr(), maxOutputLines, maxOutputBytes),
	}
	result.setData(straceResult, dns)
	return &result, nil
}

func (d *Result) setData(straceResult *strace.Result, dns *dnsanalyzer.DNSAnalyzer) {
	for _, f := range straceResult.Files() {
		d.Files = append(d.Files, fileResult{
			Path:  f.Path,
			Read:  f.Read,
			Write: f.Write,
		})
	}

	for _, s := range straceResult.Sockets() {
		d.Sockets = append(d.Sockets, socketResult{
			Address:   s.Address,
			Port:      s.Port,
			Hostnames: dns.Hostnames(s.Address),
		})
	}

	for _, c := range straceResult.Commands() {
		d.Commands = append(d.Commands, commandResult{
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
		d.DNS = append(d.DNS, c)
	}
}
