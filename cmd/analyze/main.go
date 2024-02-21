package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/featureflags"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/resultstore"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/staticanalysis"
	"github.com/ossf/package-analysis/internal/useragent"
	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/internal/worker"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

var (
	pkgName            = flag.String("package", "", "package name")
	localPkg           = flag.String("local", "", "local package path")
	ecosystem          pkgecosystem.Ecosystem
	version            = flag.String("version", "", "version")
	noPull             = flag.Bool("nopull", false, "disables pulling down sandbox images")
	imageTag           = flag.String("image-tag", "", "set image tag for analysis sandboxes")
	dynamicBucket      = flag.String("dynamic-bucket", "", "bucket path for uploading dynamic analysis results")
	staticBucket       = flag.String("static-bucket", "", "bucket path for uploading static analysis results")
	executionLogBucket = flag.String("execution-log-bucket", "", "bucket path for uploading execution log (dynamic analysis)")
	fileWritesBucket   = flag.String("file-writes-bucket", "", "bucket path for uploading file writes data (dynamic analysis)")
	analyzedPkgBucket  = flag.String("analyzed-pkg-bucket", "", "bucket path for uploading analyzed packages")
	offline            = flag.Bool("offline", false, "disables sandbox network access")
	customSandbox      = flag.String("sandbox-image", "", "override default dynamic analysis sandbox with custom image")
	customAnalysisCmd  = flag.String("analysis-command", "", "override default dynamic analysis script path (use with custom sandbox image)")
	listModes          = flag.Bool("list-modes", false, "prints out a list of available analysis modes")
	features           = flag.String("features", "", "override features that are enabled/disabled by default")
	listFeatures       = flag.Bool("list-features", false, "list available features that can be toggled")
	help               = flag.Bool("help", false, "print help on available options")
	analysisMode       = utils.CommaSeparatedFlags("mode", []string{"static", "dynamic"},
		"list of analysis modes to run, separated by commas. Use -list-modes to see available options")
)

// usageError wraps an error, to signal that the error arises from incorrect user input.
type usageError struct {
	error
}

func usagef(format string, args ...any) error {
	return usageError{fmt.Errorf(format, args...)}
}

func makeResultStores() worker.ResultStores {
	rs := worker.ResultStores{}

	if *analyzedPkgBucket != "" {
		rs.AnalyzedPackage = resultstore.New(*analyzedPkgBucket)
	}
	if *dynamicBucket != "" {
		rs.DynamicAnalysis = resultstore.New(*dynamicBucket)
	}
	if *executionLogBucket != "" {
		rs.ExecutionLog = resultstore.New(*executionLogBucket)
	}
	if *fileWritesBucket != "" {
		rs.FileWrites = resultstore.New(*fileWritesBucket)
	}
	if *staticBucket != "" {
		rs.StaticAnalysis = resultstore.New(*staticBucket)
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

func printFeatureFlags() {
	fmt.Printf("Feature List\n\n")
	fmt.Printf("%-30s %s\n", "Name", "Default")
	fmt.Printf("----------------------------------------\n")

	// print features in sorted order
	state := featureflags.State()
	sortedFeatures := maps.Keys(state)
	slices.Sort(sortedFeatures)

	// print Off/On rather than 'false' and 'true'
	stateStrings := map[bool]string{false: "Off", true: "On"}
	for _, feature := range sortedFeatures {
		fmt.Printf("%-30s %s\n", feature, stateStrings[state[feature]])
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

func dynamicAnalysis(ctx context.Context, pkg *pkgmanager.Pkg, resultStores *worker.ResultStores) {
	if !*offline {
		sandbox.InitNetwork(ctx)
	}

	sbOpts := append(worker.DynamicSandboxOptions(), makeSandboxOptions()...)

	if *customSandbox != "" {
		sbOpts = append(sbOpts, sandbox.Image(*customSandbox))
	}

	result, err := worker.RunDynamicAnalysis(ctx, pkg, sbOpts, *customAnalysisCmd)
	if err != nil {
		slog.ErrorContext(ctx, "Dynamic analysis aborted (run error)", "error", err)
		return
	}

	// this is only valid if RunDynamicAnalysis() returns nil err
	if result.LastStatus != analysis.StatusCompleted {
		slog.WarnContext(ctx, "Dynamic analysis phase did not complete successfully",
			"last_run_phase", string(result.LastRunPhase),
			"status", string(result.LastStatus))
	}

	if err := worker.SaveDynamicAnalysisData(ctx, pkg, resultStores, result.Data); err != nil {
		slog.ErrorContext(ctx, "Upload error", "error", err)
	}
}

func staticAnalysis(ctx context.Context, pkg *pkgmanager.Pkg, resultStores *worker.ResultStores) {
	if !*offline {
		sandbox.InitNetwork(ctx)
	}

	sbOpts := append(worker.StaticSandboxOptions(), makeSandboxOptions()...)

	data, status, err := worker.RunStaticAnalysis(ctx, pkg, sbOpts, staticanalysis.All)
	if err != nil {
		slog.ErrorContext(ctx, "Static analysis aborted", "error", err)
		return
	}

	slog.InfoContext(ctx, "Static analysis completed", "status", string(status))

	if err := worker.SaveStaticAnalysisData(ctx, pkg, resultStores, data); err != nil {
		slog.ErrorContext(ctx, "Upload error", "error", err)
	}
}

func run() error {
	log.Initialize(os.Getenv("LOGGER_ENV"))

	flag.TextVar(&ecosystem, "ecosystem", pkgecosystem.None, "package ecosystem. Available: "+
		strings.Join(pkgecosystem.SupportedEcosystemsStrings, ", "))

	analysisMode.InitFlag()
	flag.Parse()

	http.DefaultTransport = useragent.DefaultRoundTripper(http.DefaultTransport, "")

	if err := featureflags.Update(*features); err != nil {
		return usageError{err}
	}

	if *help {
		flag.Usage()
		return nil
	}

	if *listModes {
		printAnalysisModes()
		return nil
	}

	if *listFeatures {
		printFeatureFlags()
		return nil
	}

	if ecosystem == pkgecosystem.None {
		flag.Usage()
		return usagef("missing ecosystem")
	}

	manager := pkgmanager.Manager(ecosystem)
	if manager == nil {
		return usagef("unsupported package ecosystem %q", ecosystem)
	}

	if *pkgName == "" {
		flag.Usage()
		return usagef("missing package name")
	}

	runMode := make(map[analysis.Mode]bool)
	for _, analysisName := range analysisMode.Values {
		mode, ok := analysis.ModeFromString(strings.ToLower(analysisName))
		if !ok {
			printAnalysisModes()
			return usagef("unknown analysis mode %q", analysisName)
		}
		runMode[mode] = true
	}

	ctx := log.ContextWithAttrs(context.Background(),
		slog.Any("ecosystem", ecosystem),
	)

	slog.InfoContext(ctx, "Got request",
		slog.String("requested_name", *pkgName),
		slog.String("requested_version", *version),
	)

	pkg, err := worker.ResolvePkg(manager, *pkgName, *version, *localPkg)
	if err != nil {
		slog.ErrorContext(ctx, "Error resolving package", "error", err)
		return err
	}

	ctx = log.ContextWithAttrs(ctx,
		slog.String("name", pkg.Name()),
		slog.String("version", pkg.Version()),
	)

	slog.InfoContext(ctx, "Processing resolved package", "package_path", *localPkg)
	resultStores := makeResultStores()

	if runMode[analysis.Static] {
		slog.InfoContext(ctx, "Starting static analysis")
		staticAnalysis(ctx, pkg, &resultStores)
	}

	// dynamicAnalysis() currently panics on error, so it's last
	if runMode[analysis.Dynamic] {
		slog.InfoContext(ctx, "Starting dynamic analysis")
		dynamicAnalysis(ctx, pkg, &resultStores)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		if errors.As(err, &usageError{}) {
			fmt.Fprintf(os.Stderr, "Usage error: %v\n", err)
			os.Exit(2)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
