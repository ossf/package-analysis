package analysis

import (
	"fmt"

	"github.com/ossf/package-analysis/internal/dnsanalyzer"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/packetcapture"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/strace"
)

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

type Result struct {
	Files    []fileResult
	Sockets  []socketResult
	Commands []commandResult
}

const (
	maxIndexEntries = 10000
)

func Run(sb sandbox.Sandbox, args []string) (*Result, error) {
	log.Info("Running analysis",
		"args", args)

	log.Debug("Preparing packet capture")
	pcap, err := packetcapture.New()
	if err != nil {
		return nil, fmt.Errorf("failed to init packet capture (%w)", err)
	}

	dns := dnsanalyzer.New()
	pcap.RegisterReceiver(dns)
	if err := pcap.Start(); err != nil {
		return nil, fmt.Errorf("failed to start packet capture (%w)", err)
	}
	defer pcap.Close()

	// Run the command
	log.Debug("Running the command",
		"args", args)
	r, err := sb.Run(args...)
	if err != nil {
		return nil, fmt.Errorf("sandbox failed (%w)", err)
	}

	pcap.Close()

	// Grab the log file
	log.Debug("Parsing the strace log")
	l, err := r.Log()
	if err != nil {
		return nil, fmt.Errorf("failed to open strace log (%w)", err)
	}
	defer l.Close()

	straceResult, err := strace.Parse(l)
	if err != nil {
		return nil, fmt.Errorf("strace parsing failed (%w)", err)
	}

	result := Result{}
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

}
