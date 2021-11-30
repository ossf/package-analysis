package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/ossf/package-analysis/internal/dnsanalyzer"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/packetcapture"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/strace"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
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

type Package struct {
	Ecosystem string
	Name      string
	Version   string
}

type AnalysisResult struct {
	Package  Package
	Files    []fileResult
	Sockets  []socketResult
	Commands []commandResult
}

const (
	maxIndexEntries = 10000
)

func RunLocal(ecosystem, pkgPath, version, image, command string) *AnalysisResult {
	return run(ecosystem, pkgPath, version, image, command, []string{
		"-v", fmt.Sprintf("%s:%s", pkgPath, pkgPath),
	})
}

func RunLive(ecosystem, pkgName, version, image, command string) *AnalysisResult {
	return run(ecosystem, pkgName, version, image, command, nil)
}

func run(ecosystem, pkgName, version, image, command string, args []string) *AnalysisResult {
	log.Info("Running analysis",
		"command", command,
		"args", args)

	// Init the sandbox
	log.Debug("Init the sandbox")
	sb, err := sandbox.Init(image)
	if err != nil {
		log.Panic("Failed to init sandbox",
			"error", err)
	}

	log.Debug("Preparing packet capture")
	pcap, err := packetcapture.New()
	if err != nil {
		log.Panic("Failed to init packet capture",
			"error", err)
	}

	dns := dnsanalyzer.New()
	pcap.RegisterReceiver(dns)
	if err := pcap.Start(); err != nil {
		log.Panic("Failed to start packet capture",
			"error", err)
	}
	defer pcap.Close()

	// Run the command
	log.Debug("Running the command",
		"command", command,
		"args", args)
	r, err := sb.Run(command, args...)
	if err != nil {
		log.Panic("Command exited unsucessfully",
			"error", err)
	}

	pcap.Close()

	// Grab the log file
	log.Debug("Parsing the strace log")
	l, err := r.Log()
	if err != nil {
		log.Panic("Failed to open the log",
			"error", err)
	}
	defer l.Close()

	straceResult, err := strace.Parse(l)
	if err != nil {
		log.Panic("Failed to parse the strace",
			"error", err)
	}

	result := AnalysisResult{}
	result.setData(ecosystem, pkgName, version, straceResult, dns)
	return &result
}

func (d *AnalysisResult) setData(ecosystem, pkgName, version string, straceResult *strace.Result, dns *dnsanalyzer.DNSAnalyzer) {
	d.Package.Ecosystem = ecosystem
	d.Package.Name = pkgName
	d.Package.Version = version

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

func UploadResults(ctx context.Context, bucket, path string, result *AnalysisResult) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	bkt, err := blob.OpenBucket(ctx, bucket)
	if err != nil {
		return err
	}
	defer bkt.Close()

	filename := "results.json"
	if result.Package.Version != "" {
		filename = result.Package.Version + ".json"
	}

	uploadPath := filepath.Join(path, filename)
	log.Info("Uploading results",
		"bucket", bucket,
		"path", uploadPath)

	w, err := bkt.NewWriter(ctx, uploadPath, nil)
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return nil
}
