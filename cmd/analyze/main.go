package main

import (
	"context"
	"flag"
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
	"github.com/ossf/package-analysis/pkg/api"
)

var (
	pkgName             = flag.String("package", "", "package name")
	localPkg            = flag.String("local", "", "local package path")
	ecosystem           = flag.String("ecosystem", "", "ecosystem (npm, pypi, or rubygems)")
	version             = flag.String("version", "", "version")
	noPull              = flag.Bool("nopull", false, "disables pulling down sandbox images")
	imageTag            = flag.String("image-tag", "", "set image tag for analysis sandboxes")
	dynamicUpload       = flag.String("upload", "", "bucket path for uploading dynamic analysis results")
	staticUpload        = flag.String("upload-static", "", "bucket path for uploading static analysis results")
	uploadFileWriteInfo = flag.String("upload-file-write-info", "", "bucket path for uploading information from file writes")
	offline             = flag.Bool("offline", false, "disables sandbox network access")
	listModes           = flag.Bool("list-modes", false, "prints out a list of available analysis modes")
	help                = flag.Bool("help", false, "print help on available options")
	analysisMode        = utils.CommaSeparatedFlags("mode", "dynamic",
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

func printAnalysisModes() {
	println("Available analysis modes:")
	for _, mode := range analysis.AllModes() {
		println(mode)
	}
	println()
}

/*
makeSandboxOptions prepares options for the sandbox based on command line arguments.
In particular:
1. The image tag is always passed through. An empty tag is the same as "latest".
2. A local package is mapped into the sandbox if applicable
3. Image pulling is disabled if the "-nopull" command-line flag was used
*/
func makeSandboxOptions(mode analysis.Mode) []sandbox.Option {
	sbOpts := worker.DefaultSandboxOptions(mode, *imageTag)

	if *localPkg != "" {
		sbOpts = append(sbOpts, sandbox.Volume(*localPkg, *localPkg))
	}
	if *noPull {
		sbOpts = append(sbOpts, sandbox.NoPull())
	}
	if *offline {
		sbOpts = append(sbOpts, sandbox.Offline())
	}

	return sbOpts
}

func dynamicAnalysis(pkg *pkgecosystem.Pkg) {
	if !*offline {
		sandbox.InitNetwork()
	}

	sbOpts := makeSandboxOptions(analysis.Dynamic)

	results, lastRunPhase, lastStatus, err := worker.RunDynamicAnalysis(pkg, sbOpts)
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
		err := resultstore.New(bucket, resultstore.BasePath(path)).Save(ctx, pkg, results.FileWritesSummary)
		if err != nil {
			log.Fatal("Failed to upload file write analysis results to blobstore", "error", err)
		}
		for _, writeBufferPath := range results.FileWriteBufferPaths {
			writeBuffer, err := utils.ReadTempFile(writeBufferPath)
			if err != nil {
				log.Error("Could not read file", err)
			}
			writeBufferString := string(writeBuffer)
			writeBufferErr := resultstore.New(bucket, resultstore.BasePath(path)).SaveWriteBuffer(ctx, pkg, writeBufferString, utils.GetSHA256Hash(writeBufferString))
			if writeBufferErr != nil {
				log.Fatal(" Failed to upload file write buffer results to blobstore", "error")
			}
		}
		utils.CleanUpTempFiles()
	}

	// this is only valid if RunDynamicAnalysis() returns nil err
	if lastStatus != analysis.StatusCompleted {
		log.Fatal("Dynamic analysis phase did not complete successfully",
			"lastRunPhase", lastRunPhase,
			"status", lastStatus)
	}
}

func staticAnalysis(pkg *pkgecosystem.Pkg) {
	if !*offline {
		sandbox.InitNetwork()
	}

	sbOpts := makeSandboxOptions(analysis.Static)

	results, status, err := worker.RunStaticAnalysis(pkg, sbOpts, staticanalysis.All)
	if err != nil {
		log.Fatal("Static analysis aborted", "error", err)
	}

	log.Info("Static analysis completed", "status", status)

	ctx := context.Background()
	if *staticUpload != "" {
		bucket, path := parseBucketPath(*staticUpload)
		err := resultstore.New(bucket, resultstore.BasePath(path)).Save(ctx, pkg, results)
		if err != nil {
			log.Fatal("Failed to upload static results to blobstore", "error", err)
		}
	}
}

func main() {
	log.Initialize(os.Getenv("LOGGER_ENV"))

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

	if *ecosystem == "" {
		flag.Usage()
		return
	}

	manager := pkgecosystem.Manager(api.Ecosystem(*ecosystem))
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
