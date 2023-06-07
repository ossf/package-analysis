package worker

import (
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/dynamicanalysis"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
)

var nonSpaceControlChars = regexp.MustCompile("[\x00-\x08\x0b-\x1f\x7f]")

func retrieveExecutionLog(sb sandbox.Sandbox) (string, error) {
	// retrieve execution log back to host
	executionLogDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	defer os.RemoveAll(executionLogDir)
	executionLogPath := path.Join(executionLogDir, "execution.log")

	// if the copy fails, it could be that the execution log is not actually present.
	// For now, we'll just log the error and otherwise ignore it
	if err := sb.CopyToHost("/execution.log", executionLogPath); err != nil {
		log.Warn("Could not copy execution log from sandbox", "error", err)
		return "", nil
	}

	logData, err := os.ReadFile(executionLogPath)
	if err != nil {
		return "", err
	}

	// remove control characters except tab (\x09) and newline (\x0A)
	processedLog := nonSpaceControlChars.ReplaceAllLiteral(logData, []byte{})
	log.Info("Read execution log", "rawLength", len(logData), "processedLength", len(processedLog))

	return string(processedLog), nil
}

/*
RunDynamicAnalysis runs dynamic analysis on the given package in the sandbox
provided, across all phases (e.g. import, install) valid in the package ecosystem.
Status and errors are logged to stdout. There are 4 return values:

DynamicAnalysisResults: Map of each successfully run phase to a summary of
the corresponding dynamic analysis result. This summary has two parts:
1. StraceSummary: information about system calls performed by the process
2. FileWrites: list of files which were written to and counts of bytes written

Note, if error is not nil, then results[lastRunPhase] is nil.

RunPhase: the last phase that was run. If error is non-nil, this phase did not
successfully complete, and the results for this phase are not recorded.
Otherwise, the results contain data for this phase, even in cases where the
sandboxed process terminated abnormally.

Status: the status of the last run phase if it completed without error, else empty

error: Any error that occurred in the runtime/sandbox infrastructure.
This does not include errors caused by the package under analysis.
*/

func RunDynamicAnalysis(pkg *pkgmanager.Pkg, sbOpts []sandbox.Option) (analysisrun.DynamicAnalysisResults, analysisrun.DynamicPhase, analysis.Status, error) {

	var beforeDynamic runtime.MemStats
	runtime.ReadMemStats(&beforeDynamic)
	log.Info("Memory Stats, heap usage before dynamic analysis",
		log.Label("ecosystem", pkg.EcosystemName()),
		"name", pkg.Name(),
		"version", pkg.Version(),
		"heap_usage_before_dynamic_analysis", strconv.FormatUint(beforeDynamic.Alloc, 10),
	)

	results := analysisrun.DynamicAnalysisResults{
		StraceSummary:      make(analysisrun.DynamicAnalysisStraceSummary),
		FileWritesSummary:  make(analysisrun.DynamicAnalysisFileWritesSummary),
		FileWriteBufferIds: make(analysisrun.DynamicAnalysisFileWriteBufferIds),
	}

	sb := sandbox.New(pkg.Manager().DynamicAnalysisImage(), sbOpts...)

	defer func() {
		if err := sb.Clean(); err != nil {
			log.Error("error cleaning up sandbox", "error", err)
		}
	}()

	var lastRunPhase analysisrun.DynamicPhase
	var lastStatus analysis.Status
	var lastError error
	for _, phase := range pkg.Manager().DynamicPhases() {
		startTime := time.Now()
		phaseResult, err := dynamicanalysis.Run(sb, pkg.Command(phase))
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

		results.StraceSummary[phase] = &phaseResult.StraceSummary
		results.FileWritesSummary[phase] = &phaseResult.FileWritesSummary
		lastStatus = phaseResult.StraceSummary.Status
		results.FileWriteBufferIds[phase] = phaseResult.FileWriteBufferIds

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

	if lastError != nil {
		LogDynamicAnalysisError(pkg, lastRunPhase, lastError)
		return results, lastRunPhase, lastStatus, lastError
	}

	LogDynamicAnalysisResult(pkg, lastRunPhase, lastStatus)

	executionLog, err := retrieveExecutionLog(sb)
	if err != nil {
		// don't return this error, just log it
		log.Error("Error retrieving execution log", "error", err)
	} else {
		results.ExecutionLog = analysisrun.DynamicAnalysisExecutionLog(executionLog)
	}

	return results, lastRunPhase, lastStatus, nil
}
