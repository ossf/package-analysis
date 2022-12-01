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
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/resultstore"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/staticanalysis"
	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/internal/worker"
)

var (
	pkgName       = flag.String("package", "", "package name")
	localPkg      = flag.String("local", "", "local package path")
	ecosystem     = flag.String("ecosystem", "", "ecosystem (npm, pypi, or rubygems)")
	version       = flag.String("version", "", "version")
	noPull        = flag.Bool("nopull", false, "disables pulling down sandbox images")
	dynamicUpload = flag.String("upload", "", "bucket path for uploading dynamic analysis results")
	imageTag      = flag.String("image-tag", "", "set image tag for analysis sandboxes")
	staticUpload  = flag.String("upload-static", "", "bucket path for uploading static analysis results")
	listModes     = flag.Bool("list-modes", false, "prints out a list of available analysis modes")
	analysisMode  = utils.CommaSeparatedFlags("analysis-mode", "dynamic",
		"single or comma separated list of analysis modes to run. Use -list-modes to see available options")
	uploadFileWriteInfo = flag.String("upload-file-write-info", "", "bucket path for uploading information from file writes")
)

func parseBucketPath(path string) (string, string) {
	parsed, err := url.Parse(path)
	if err != nil {
		log.Panic("Failed to parse bucket path",
			"path", path)
	}

	return parsed.Scheme + "://" + parsed.Host, parsed.Path
}

func cleanupSandbox(sb sandbox.Sandbox) {
	err := sb.Clean()
	if err != nil {
		log.Error("error cleaning up sandbox", "error", err)
	}
}

func printAnalysisModes() {
	println("Available analysis modes:")
	for _, mode := range analysis.AllModes() {
		println(mode)
	}
	println()
}

/*
makeSandboxOptions prepares options for the sandbox based on command line arguments:
1. Always pass through the tag. An empty tag is the same as "latest".
2. Respect the "-nopull" option.
3. Ensure any local package is mapped through.
*/
func makeSandboxOptions(mode analysis.Mode) []sandbox.Option {
	sbOpts := worker.DefaultSandboxOptions(mode, *imageTag)

	if *localPkg != "" {
		sbOpts = append(sbOpts, sandbox.Volume(*localPkg, *localPkg))
	}
	if *noPull {
		sbOpts = append(sbOpts, sandbox.NoPull())
	}

	return sbOpts
}

func dynamicAnalysis(pkg *pkgecosystem.Pkg) {
	sandbox.InitEnv()
	sbOpts := makeSandboxOptions(analysis.Dynamic)
	sb := sandbox.New(pkg.Manager().DynamicAnalysisImage(), sbOpts...)
	defer cleanupSandbox(sb)

	results, lastRunPhase, err := worker.RunDynamicAnalysis(sb, pkg)
	if err != nil {
		log.Fatal("Dynamic analysis aborted (run error)", "error", err)
	}

	ctx := context.Background()
	if *dynamicUpload != "" {
		bucket, path := parseBucketPath(*dynamicUpload)
		err := resultstore.New(bucket, resultstore.BasePath(path)).Save(ctx, pkg, results.StraceSummary)
		if err != nil {
			log.Fatal("Failed to upload dynamic analysis results to blobstore",
				"error", err)
		}
	}

	if *uploadFileWriteInfo != "" {
		bucket, path := parseBucketPath(*uploadFileWriteInfo)
		err := resultstore.New(bucket, resultstore.BasePath(path)).Save(ctx, pkg, results.FileWrites)
		if err != nil {
			log.Fatal("Failed to upload file write analysis results to blobstore", "error", err)
		}
	}

	// this is only valid if RunDynamicAnalysis() returns nil err
	lastStatus := results.StraceSummary[lastRunPhase].Status
	if lastStatus != analysis.StatusCompleted {
		log.Fatal("Dynamic analysis phase did not complete successfully",
			"lastRunPhase", lastRunPhase,
			"status", lastStatus)
	}
}

func staticAnalysis(pkg *pkgecosystem.Pkg) {
	sandbox.InitEnv()
	sbOpts := makeSandboxOptions(analysis.Static)

	image := "gcr.io/ossf-malware-analysis/static-analysis"

	sb := sandbox.New(image, sbOpts...)
	defer cleanupSandbox(sb)

	results, err := worker.RunStaticAnalyses(sb, pkg, staticanalysis.ObfuscationDetection)
	if err != nil {
		log.Fatal("Static analysis aborted", "error", err)
	}

	for task, result := range results {
		fmt.Printf("%s result\n%s\n", task, result)
	}

	ctx := context.Background()
	if *staticUpload != "" {
		bucket, path := parseBucketPath(*staticUpload)
		err := resultstore.New(bucket, resultstore.BasePath(path)).Save(ctx, pkg, results)
		if err != nil {
			log.Fatal("Failed to upload static results to blobstore",
				"error", err)
		}
	}
}

func main() {
	log.Initalize(os.Getenv("LOGGER_ENV"))

	analysisMode.InitFlag()
	flag.Parse()

	if *listModes {
		printAnalysisModes()
		return
	}

	if *ecosystem == "" {
		flag.Usage()
		return
	}

	manager := pkgecosystem.Manager(pkgecosystem.Ecosystem(*ecosystem))
	if manager == nil {
		log.Panic("Unsupported pkg manager",
			log.Label("ecosystem", *ecosystem))
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

	worker.LogRequest(*ecosystem, *pkgName, *version, *localPkg, "")

	pkg, err := worker.ResolvePkg(manager, *pkgName, *version, *localPkg)
	if err != nil {
		log.Panic("Error resolving package",
			log.Label("ecosystem", *ecosystem),
			log.Label("name", *pkgName),
			"error", err)
	}

	if runMode[analysis.Static] {
		log.Info("Starting static analysis")
		staticAnalysis(pkg)
	}

	// dynamicAnalysis() currently panics on error, so it's last
	if runMode[analysis.Dynamic] {
		log.Info("Starting dynamic analysis")
		dynamicAnalysis(pkg)
	}
}
