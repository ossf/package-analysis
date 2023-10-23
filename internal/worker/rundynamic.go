package worker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/dynamicanalysis"
	"github.com/ossf/package-analysis/internal/featureflags"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

// defaultDynamicAnalysisImage is container image name of the default dynamic analysis sandbox
const defaultDynamicAnalysisImage = "gcr.io/ossf-malware-analysis/dynamic-analysis"

/*
DynamicAnalysisResult holds all the results from RunDynamicAnalysis

AnalysisData: Map of each successfully run phase to a summary of
the corresponding dynamic analysis result. This summary has two parts:
1. StraceSummary: information about system calls performed by the process
2. FileWrites: list of files which were written to and counts of bytes written

Note, if error is not nil, then results[lastRunPhase] is nil.

LastRunPhase: the last phase that was run. If error is non-nil, this phase did not
successfully complete, and the results for this phase are not recorded.
Otherwise, the results contain data for this phase, even in cases where the
sandboxed process terminated abnormally.

Status: the status of the last run phase if it completed without error, else empty
*/

type DynamicAnalysisResult struct {
	AnalysisData analysisrun.DynamicAnalysisResults
	LastRunPhase analysisrun.DynamicPhase
	LastStatus   analysis.Status
}

func dynamicPhases(ecosystem pkgecosystem.Ecosystem) []analysisrun.DynamicPhase {
	phases := analysisrun.DefaultDynamicPhases()

	// currently, the execute phase is only supported for python analysis
	executePhaseSupported := map[pkgecosystem.Ecosystem]struct{}{
		pkgecosystem.PyPI: {},
	}

	if featureflags.CodeExecution.Enabled() {
		if _, supported := executePhaseSupported[ecosystem]; supported {
			phases = append(phases, analysisrun.DynamicPhaseExecute)
		}
	}

	return phases
}

/*
RunDynamicAnalysis runs dynamic analysis on the given package across the phases
valid in the package ecosystem (e.g. import, install), in a sandbox created
using the provided options. The options must specify the sandbox image to use.

analysisCmd is an optional argument used to override the default command run
inside the sandbox to perform the analysis. It must support the interface
described under "Adding a new Runtime Analysis script" in sandboxes/README.md

All data and status relating to analysis (including errors produced by invalid packages)
is returned in the DynamicAnalysisResult struct. Status and errors are also logged to stdout.

The returned error holds any error that occurred in the runtime/sandbox infrastructure,
excluding from within the analysis itself. In other words, it does not include errors
produced by the package under analysis.
*/
func RunDynamicAnalysis(ctx context.Context, pkg *pkgmanager.Pkg, sbOpts []sandbox.Option, analysisCmd string) (DynamicAnalysisResult, error) {
	ctx = log.ContextWithAttrs(ctx, slog.String("mode", "dynamic"))

	var beforeDynamic runtime.MemStats
	runtime.ReadMemStats(&beforeDynamic)
	slog.InfoContext(ctx, "Memory Stats, heap usage before dynamic analysis",
		"heap_usage_before_dynamic_analysis", strconv.FormatUint(beforeDynamic.Alloc, 10),
	)

	if analysisCmd == "" {
		analysisCmd = dynamicanalysis.DefaultCommand(pkg.Ecosystem())
	}

	sb := sandbox.New(sbOpts...)

	defer func() {
		if err := sb.Clean(ctx); err != nil {
			slog.ErrorContext(ctx, "Error cleaning up sandbox", "error", err)
		}
	}()

	// initialise sandbox before copy/run
	if err := sb.Init(ctx); err != nil {
		LogDynamicAnalysisError(ctx, pkg, "", err)
		return DynamicAnalysisResult{}, err
	}

	result := DynamicAnalysisResult{
		AnalysisData: analysisrun.DynamicAnalysisResults{
			StraceSummary:      make(analysisrun.DynamicAnalysisStraceSummary),
			FileWritesSummary:  make(analysisrun.DynamicAnalysisFileWritesSummary),
			FileWriteBufferIds: make(analysisrun.DynamicAnalysisFileWriteBufferIds),
		},
	}

	// lastError holds the error that occurred in the most recently run dynamic analysis phase.
	// This is not a part of the result because a non-nil value means that the error originated
	// from our code, as opposed to the package under analysis
	var lastError error

	for _, phase := range dynamicPhases(pkg.Ecosystem()) {
		if err := runDynamicAnalysisPhase(ctx, pkg, sb, analysisCmd, phase, &result); err != nil {
			// Error when trying to actually run; don't record the result for this phase
			// or attempt subsequent phases
			result.LastStatus = ""
			lastError = err
			break
		}

		if result.LastStatus != analysis.StatusCompleted {
			// Error caused by an issue with the package (probably).
			// Don't continue with phases if this one did not complete successfully.
			break
		}
	}

	var afterDynamic runtime.MemStats
	runtime.ReadMemStats(&afterDynamic)
	slog.InfoContext(ctx, "Memory Stats, heap usage after dynamic analysis",
		"heap_usage_after_dynamic_analysis", strconv.FormatUint(afterDynamic.Alloc, 10))

	if lastError != nil {
		LogDynamicAnalysisError(ctx, pkg, result.LastRunPhase, lastError)
		return result, lastError
	}

	LogDynamicAnalysisResult(ctx, pkg, result.LastRunPhase, result.LastStatus)

	return result, nil
}

// openStraceDebugLogFile creates and returns the file to be used for debug logging of strace parsing
// during a dynamic analysis phase. The file is created with the given filename in log.StraceDebugLogDir.
// It is truncated on open (so a unique name per analysis phase should be used) and is the caller's
// responsibility to close. If strace debug logging is disabled, or some error occurs during creation,
// a nil file pointer is returned, and nothing more need be done by the caller.
func openStraceDebugLogFile(ctx context.Context, name string) *os.File {
	if !featureflags.StraceDebugLogging.Enabled() {
		return nil
	}

	var logDir = log.StraceDebugLogDir
	if err := os.MkdirAll(logDir, 0o777); err != nil {
		slog.WarnContext(ctx, "could not create directory for strace debug logs", "path", logDir, "error", err)
	}

	logPath := filepath.Join(logDir, name)
	if logFile, err := os.Create(logPath); err != nil {
		slog.WarnContext(ctx, "could not create strace debug log file", "path", logPath, "error", err)
		return nil
	} else {
		return logFile
	}
}

func straceDebugLogFilename(pkg *pkgmanager.Pkg, phase analysisrun.DynamicPhase) string {
	filename := fmt.Sprintf("%s-%s", pkg.Ecosystem(), pkg.Name())
	if pkg.Version() != "" {
		filename += "-" + pkg.Version()
	}
	filename += fmt.Sprintf("-%s-strace.log", phase)

	// Protect against e.g. a package name that contains a slash.
	// This may cause name collisions, but it's probably fine for a debug log
	return strings.ReplaceAll(filename, string(os.PathSeparator), "-")
}

func runDynamicAnalysisPhase(ctx context.Context, pkg *pkgmanager.Pkg, sb sandbox.Sandbox, analysisCmd string, phase analysisrun.DynamicPhase, result *DynamicAnalysisResult) error {
	phaseCtx := log.ContextWithAttrs(ctx, log.Label("phase", string(phase)))
	startTime := time.Now()
	args := dynamicanalysis.MakeAnalysisArgs(pkg, phase)

	straceLogger := slog.New(slog.NewTextHandler(io.Discard, nil)) // default is nop logger
	if logFile := openStraceDebugLogFile(phaseCtx, straceDebugLogFilename(pkg, phase)); logFile != nil {
		slog.InfoContext(phaseCtx, "strace debug logging enabled")
		defer logFile.Close()

		enableDebug := &slog.HandlerOptions{Level: slog.LevelDebug}
		straceLogger = slog.New(log.NewContextLogHandler(slog.NewTextHandler(logFile, enableDebug)))
		straceLogger.InfoContext(phaseCtx, "running dynamic analysis")
	}

	phaseResult, err := dynamicanalysis.Run(phaseCtx, sb, analysisCmd, args, straceLogger)
	result.LastRunPhase = phase
	runDuration := time.Since(startTime)
	slog.InfoContext(phaseCtx, "Dynamic analysis phase finished",
		"error", err,
		"dynamic_analysis_phase_duration", runDuration,
	)

	if err != nil {
		return err
	}

	result.AnalysisData.StraceSummary[phase] = &phaseResult.StraceSummary
	result.AnalysisData.FileWritesSummary[phase] = &phaseResult.FileWritesSummary
	result.AnalysisData.FileWriteBufferIds[phase] = phaseResult.FileWriteBufferIds
	result.LastStatus = phaseResult.StraceSummary.Status

	if phase == analysisrun.DynamicPhaseExecute {
		executionLog, err := retrieveExecutionLog(ctx, sb)
		if err != nil {
			// don't return this error, just log it
			slog.ErrorContext(ctx, "Error retrieving execution log", "error", err)
		} else {
			result.AnalysisData.ExecutionLog = analysisrun.DynamicAnalysisExecutionLog(executionLog)
		}
	}

	return nil
}
