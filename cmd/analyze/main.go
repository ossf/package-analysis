package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/resultstore"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/staticanalysis"
	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/internal/worker"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

var (
	pkgName             = flag.String("package", "", "package name")
	localPkg            = flag.String("local", "", "local package path")
	ecosystem           pkgecosystem.Ecosystem
	version             = flag.String("version", "", "version")
	noPull              = flag.Bool("nopull", false, "disables pulling down sandbox images")
	imageTag            = flag.String("image-tag", "", "set image tag for analysis sandboxes")
	dynamicUpload       = flag.String("upload", "", "bucket path for uploading dynamic analysis results")
	staticUpload        = flag.String("upload-static", "", "bucket path for uploading static analysis results")
	uploadFileWriteInfo = flag.String("upload-file-write-info", "", "bucket path for uploading information from file writes")
	offline             = flag.Bool("offline", false, "disables sandbox network access")
	customSandbox       = flag.String("sandbox-image", "", "override default dynamic analysis sandbox with custom image")
	customAnalysisCmd   = flag.String("analysis-command", "", "override default dynamic analysis script path (use with custom sandbox image)")
	listModes           = flag.Bool("list-modes", false, "prints out a list of available analysis modes")
	help                = flag.Bool("help", false, "print help on available options")
	analysisMode        = utils.CommaSeparatedFlags("mode", []string{"static", "dynamic"},
		"list of analysis modes to run, separated by commas. Use -list-modes to see available options")
)

func parseBucketPath(path string) (string, string) {
	parsed, err := url.Parse(path)
	if err != nil {
		log.Panic("Failed to parse bucket path",
			"path", path)
	}

	return parsed.Scheme + "://" + parsed.Host, parsed.Path
}

func makeResultStores() worker.ResultStores {
	rs := worker.ResultStores{}

	if *dynamicUpload != "" {
		bucket, path := parseBucketPath(*dynamicUpload)
		rs.DynamicAnalysis = resultstore.New(bucket, resultstore.BasePath(path))
	}
	if *staticUpload != "" {
		bucket, path := parseBucketPath(*staticUpload)
		rs.StaticAnalysis = resultstore.New(bucket, resultstore.BasePath(path))
	}
	if *uploadFileWriteInfo != "" {
		bucket, path := parseBucketPath(*uploadFileWriteInfo)
		rs.FileWrites = resultstore.New(bucket, resultstore.BasePath(path))
	}

	return rs
}

func printAnalysisModes() {
	fmt.Println("Available analysis modes:")
	for _, mode := range analysis.AllModes() {
		fmt.Println(mode)
	}
	fmt.Println()
}

// makeSandboxOptions prepares options for the sandbox based on command line arguments.
//
// In particular:
//
//  1. The image tag is always passed through. An empty tag is the same as "latest".
//  2. A local package is mapped into the sandbox if applicable.
//  3. Image pulling is disabled if the "-nopull" command-line flag was used.
func makeSandboxOptions() []sandbox.Option {
	sbOpts := []sandbox.Option{sandbox.Tag(*imageTag)}

	if *localPkg != "" {
		sbOpts = append(sbOpts, sandbox.Copy(*localPkg, *localPkg))
	}
	if *noPull {
		sbOpts = append(sbOpts, sandbox.NoPull())
	}
	if *offline {
		sbOpts = append(sbOpts, sandbox.Offline())
	}

	return sbOpts
}

func dynamicAnalysis(pkg *pkgmanager.Pkg, resultStores worker.ResultStores) {
	if !*offline {
		sandbox.InitNetwork()
	}

	sbOpts := append(worker.DynamicSandboxOptions(pkg.Ecosystem()), makeSandboxOptions()...)

	if *customSandbox != "" {
		sbOpts = append(sbOpts, sandbox.Image(*customSandbox))
	}
	if *customAnalysisCmd != "" {
		sbOpts = append(sbOpts, sandbox.Command(*customAnalysisCmd))
	}

	result, err := worker.RunDynamicAnalysis(pkg, sbOpts)
	if err != nil {
		log.Error("Dynamic analysis aborted (run error)", "error", err)
		return
	}

	// this is only valid if RunDynamicAnalysis() returns nil err
	if result.LastStatus != analysis.StatusCompleted {
		log.Warn("Dynamic analysis phase did not complete successfully",
			"lastRunPhase", result.LastRunPhase,
			"status", result.LastStatus)
	}

	ctx := context.Background()
	if err := worker.SaveDynamicAnalysisData(ctx, pkg, resultStores, result.Data); err != nil {
		log.Error("Upload error", "error", err)
	}
}

func staticAnalysis(pkg *pkgmanager.Pkg, resultStores worker.ResultStores) {
	if !*offline {
		sandbox.InitNetwork()
	}

	sbOpts := append(worker.StaticSandboxOptions(), makeSandboxOptions()...)

	data, status, err := worker.RunStaticAnalysis(pkg, sbOpts, staticanalysis.All)
	if err != nil {
		log.Error("Static analysis aborted", "error", err)
		return
	}

	log.Info("Static analysis completed", "status", status)

	ctx := context.Background()
	if err := worker.SaveStaticAnalysisData(ctx, pkg, resultStores, data); err != nil {
		log.Error("Upload error", "error", err)
	}
}

func main() {
	log.Initialize(os.Getenv("LOGGER_ENV"))

	flag.TextVar(&ecosystem, "ecosystem", pkgecosystem.None, fmt.Sprintf("package ecosystem. Can be %s", pkgecosystem.SupportedEcosystemsStrings))

	analysisMode.InitFlag()
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *listModes {
		printAnalysisModes()
		return
	}

	if ecosystem == pkgecosystem.None {
		flag.Usage()
		return
	}

	manager := pkgmanager.Manager(ecosystem)
	if manager == nil {
		log.Panic("Unsupported pkg manager",
			log.Label("ecosystem", string(ecosystem)))
	}

	if *pkgName == "" {
		flag.Usage()
		return
	}

	runMode := make(map[analysis.Mode]bool)
	for _, analysisName := range analysisMode.Values {
		mode, ok := analysis.ModeFromString(strings.ToLower(analysisName))
		if !ok {
			log.Error("Unknown analysis mode: " + analysisName)
			printAnalysisModes()
			return
		}
		runMode[mode] = true
	}

	worker.LogRequest(ecosystem, *pkgName, *version, *localPkg, "")

	pkg, err := worker.ResolvePkg(manager, *pkgName, *version, *localPkg)
	if err != nil {
		log.Panic("Error resolving package",
			log.Label("ecosystem", ecosystem.String()),
			"name", *pkgName,
			"error", err)
	}

	resultStores := makeResultStores()

	if runMode[analysis.Static] {
		log.Info("Starting static analysis")
		staticAnalysis(pkg, resultStores)
	}

	// dynamicAnalysis() currently panics on error, so it's last
	if runMode[analysis.Dynamic] {
		log.Info("Starting dynamic analysis")
		dynamicAnalysis(pkg, resultStores)
	}
}
