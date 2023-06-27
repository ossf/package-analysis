package worker

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
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

// sandboxExecutionLogPath is the absolute path of the execution log file
// inside the sandbox. The file is used for code execution feature.
const sandboxExecutionLogPath = "/execution.log"

var nonSpaceControlChars = regexp.MustCompile("[\x00-\x08\x0b-\x1f\x7f]")

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

func shouldEnableCodeExecution(ecosystem pkgecosystem.Ecosystem) bool {
	if !featureflags.CodeExecution.Enabled() {
		return false
	}

	switch ecosystem {
	case pkgecosystem.PyPI:
		return true
	default:
		return false
	}
}

func enableCodeExecution(sb sandbox.Sandbox) error {
	// Create empty execution log file and copy to sandbox, to enable code execution feature
	tempFile, err := os.CreateTemp("", "")
	if err != nil {
		return fmt.Errorf("could not create execution log file in host: %w", err)
	}

	// file wasn't written to so don't worry too much about close errors
	_ = tempFile.Close()
	tempPath := tempFile.Name()

	if err := sb.CopyIntoSandbox(tempPath, sandboxExecutionLogPath); err != nil {
		return fmt.Errorf("could not copy execution log file to sandbox: %w", err)
	}

	// file wasn't written to so don't worry too much about remove errors
	_ = os.Remove(tempPath)

	return nil
}

// retrieveExecutionLog copies the execution log back from the sandbox
// to the host, so it can be included in the dynamic analysis results.
// To mitigate tampering of the file, all control characters except tab
// and newline are stripped from the file.
func retrieveExecutionLog(sb sandbox.Sandbox) (string, error) {
	executionLogDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	defer os.RemoveAll(executionLogDir)
	hostExecutionLogPath := filepath.Join(executionLogDir, "execution.log")

	// if the copy fails, it could be that the execution log is not actually present.
	// For now, we'll just log the error and otherwise ignore it
	if err := sb.CopyBackToHost(sandboxExecutionLogPath, hostExecutionLogPath); err != nil {
		log.Warn("Could not copy execution log from sandbox", "error", err)
		return "", nil
	}

	logData, err := os.ReadFile(hostExecutionLogPath)
	if err != nil {
		return "", err
	}

	// remove control characters except tab (\x09) and newline (\x0A)
	processedLog := nonSpaceControlChars.ReplaceAllLiteral(logData, []byte{})
	log.Info("Read execution log", "rawLength", len(logData), "processedLength", len(processedLog))

	return string(processedLog), nil
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
func RunDynamicAnalysis(pkg *pkgmanager.Pkg, sbOpts []sandbox.Option, analysisCmd string) (DynamicAnalysisResult, error) {
	var beforeDynamic runtime.MemStats
	runtime.ReadMemStats(&beforeDynamic)
	log.Info("Memory Stats, heap usage before dynamic analysis",
		log.Label("ecosystem", pkg.EcosystemName()),
		"name", pkg.Name(),
		"version", pkg.Version(),
		"heap_usage_before_dynamic_analysis", strconv.FormatUint(beforeDynamic.Alloc, 10),
	)

	if analysisCmd == "" {
		analysisCmd = dynamicanalysis.DefaultCommand(pkg.Ecosystem())
	}

	data := analysisrun.DynamicAnalysisResults{
		StraceSummary:      make(analysisrun.DynamicAnalysisStraceSummary),
		FileWritesSummary:  make(analysisrun.DynamicAnalysisFileWritesSummary),
		FileWriteBufferIds: make(analysisrun.DynamicAnalysisFileWriteBufferIds),
	}

	sb := sandbox.New(sbOpts...)

	defer func() {
		if err := sb.Clean(); err != nil {
			log.Error("error cleaning up sandbox", "error", err)
		}
	}()

	codeExecutionEnabled := false
	if shouldEnableCodeExecution(pkg.Ecosystem()) {
		if err := enableCodeExecution(sb); err != nil {
			log.Error("code execution disabled due to error", "error", err)
		} else {
			codeExecutionEnabled = true
		}
	}

	var lastRunPhase analysisrun.DynamicPhase
	var lastStatus analysis.Status
	var lastError error
	for _, phase := range analysisrun.DefaultDynamicPhases() {
		startTime := time.Now()
		args := dynamicanalysis.MakeAnalysisArgs(pkg, phase)
		phaseResult, err := dynamicanalysis.Run(sb, analysisCmd, args)
		lastRunPhase = phase

		runDuration := time.Since(startTime)
		log.Info("Dynamic analysis phase finished",
			log.Label("ecosystem", pkg.EcosystemName()),
			"name", pkg.Name(),
			"version", pkg.Version(),
			log.Label("phase", string(phase)),
			"error", err,
			"dynamic_analysis_phase_duration", runDuration,
		)

		if err != nil {
			// Error when trying to actually run; don't record the result for this phase
			// or attempt subsequent phases
			lastStatus = ""
			lastError = err
			break
		}

		data.StraceSummary[phase] = &phaseResult.StraceSummary
		data.FileWritesSummary[phase] = &phaseResult.FileWritesSummary
		lastStatus = phaseResult.StraceSummary.Status
		data.FileWriteBufferIds[phase] = phaseResult.FileWriteBufferIds

		if lastStatus != analysis.StatusCompleted {
			// Error caused by an issue with the package (probably).
			// Don't continue with phases if this one did not complete successfully.
			break
		}
	}

	var afterDynamic runtime.MemStats
	runtime.ReadMemStats(&afterDynamic)
	log.Info("Memory Stats, heap usage after dynamic analysis",
		log.Label("ecosystem", pkg.EcosystemName()),
		"name", pkg.Name(),
		"version", pkg.Version(),
		"heap_usage_after_dynamic_analysis", strconv.FormatUint(afterDynamic.Alloc, 10))

	analysisResult := DynamicAnalysisResult{
		AnalysisData: data,
		LastRunPhase: lastRunPhase,
		LastStatus:   lastStatus,
	}

	if lastError != nil {
		LogDynamicAnalysisError(pkg, lastRunPhase, lastError)
		return analysisResult, lastError
	}

	LogDynamicAnalysisResult(pkg, lastRunPhase, lastStatus)

	if !codeExecutionEnabled {
		// nothing more to do
		return analysisResult, nil
	}

	executionLog, err := retrieveExecutionLog(sb)
	if err != nil {
		// don't return this error, just log it
		log.Error("Error retrieving execution log", "error", err)
	} else {
		data.ExecutionLog = analysisrun.DynamicAnalysisExecutionLog(executionLog)
	}

	return analysisResult, nil
}
