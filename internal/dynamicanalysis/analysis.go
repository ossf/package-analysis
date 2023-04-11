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
	"io/ioutil"
	"os/exec"
	"strings"
	"syscall"
)

const (
	maxOutputLines = 20
	maxOutputBytes = 4 * 1024
)

type Result struct {
	StraceSummary analysisrun.StraceSummary
	FileWrites    analysisrun.FileWrites
	URLs          []string
}

var resultError = &Result{
	StraceSummary: analysisrun.StraceSummary{
		Status: analysis.StatusErrorOther,
	},
}

func ParseURLsFromSSlStripOutput(content []byte) []string {
	var result []string
	ret := strings.Split(string(content), "\n")

	for _, value := range ret {
		lineParts := strings.Split(value, " ")
		if len(lineParts) < 11 {
			continue
		}

		schema := lineParts[3]
		host := lineParts[8]
		path := lineParts[10]
		result = append(result, schema+"://"+host+path)
	}
	return result
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

	log.Debug("Reroute all http traffic through sslsplit")
	iptables := exec.Command("iptables", "-t", "nat", "-A", "PREROUTING", "-i", "cni-analysis", "-p", "tcp", "--dport", "80", "-j", "REDIRECT", "--to-port", "8081")
	err := iptables.Start()

	if err != nil {
		log.Fatal(err.Error())
	}
	err = iptables.Wait()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Debug("Reroute all https traffic through sslsplit")
	iptables = exec.Command("iptables", "-t", "nat", "-A", "PREROUTING", "-i", "cni-analysis", "-p", "tcp", "--dport", "443", "-j", "REDIRECT", "--to-port", "8080")
	err = iptables.Start()

	if err != nil {
		log.Fatal(err.Error())
	}
	err = iptables.Wait()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Debug("starting sslsplit")
	sslsplit := exec.Command("sslsplit", "-d", "-L", "/tmp/ssl.flow", "-l", "/tmp/sslLinks.flow", "-k", "/proxy/certs/ca.pem", "-c", "/proxy/certs/ca.crt", "http", "0.0.0.0", "8081", "https", "0.0.0.0", "8080")
	err = sslsplit.Start()

	if err != nil {
		log.Fatal(err.Error())
	}

	r, err := sb.Run(args...)
	if err != nil {
		return resultError, fmt.Errorf("sandbox failed (%w)", err)
	}

	log.Debug("stopping sslsplit")
	err = sslsplit.Process.Signal(syscall.SIGINT)
	if err != nil {
		return nil, err
	}

	log.Debug("reading sslsplit results")
	body, err1 := ioutil.ReadFile("/tmp/sslLinks.flow")
	if err1 != nil {
		log.Fatal("unable to read file: %v", err1)
	}

	urls := ParseURLsFromSSlStripOutput(body)

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
	analysisResult.setData(straceResult, dns, urls)
	return &analysisResult, nil
}

func (d *Result) setData(straceResult *strace.Result, dns *dnsanalyzer.DNSAnalyzer, sslSplitResult []string) {
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
					BytesWritten: wi.BytesWritten,
				})
			}
			d.FileWrites = append(d.FileWrites, w)
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
	d.URLs = sslSplitResult
}
