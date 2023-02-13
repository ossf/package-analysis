package worker

import (
	"time"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/dynamicanalysis"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/pkg/api"
	"github.com/ossf/package-analysis/pkg/result"
)

var combinedDynamicAnalyisImage = "gcr.io/ossf-malware-analysis/dynamic-analysis"

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

func RunDynamicAnalysis(pkg *pkgecosystem.Pkg, sbOpts []sandbox.Option, useCombinedSandbox bool) (result.DynamicAnalysisResults, api.RunPhase, analysis.Status, error) {
	results := result.DynamicAnalysisResults{
		StraceSummary: make(result.DynamicAnalysisStraceSummary),
		FileWrites:    make(result.DynamicAnalysisFileWrites),
	}

	var image string
	if useCombinedSandbox {
		image = combinedDynamicAnalyisImage
	} else {
		image = pkg.Manager().DynamicAnalysisImage()
	}
	sb := sandbox.New(image, sbOpts...)

	defer func() {
		if err := sb.Clean(); err != nil {
			log.Error("error cleaning up sandbox", "error", err)
		}
	}()

	var lastRunPhase api.RunPhase
	var lastStatus analysis.Status
	var lastError error
	for _, phase := range pkg.Manager().RunPhases() {
		startTime := time.Now()
		phaseResult, err := dynamicanalysis.Run(sb, pkg.Command(phase, useCombinedSandbox))
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
		results.FileWrites[phase] = &phaseResult.FileWrites
		lastStatus = phaseResult.StraceSummary.Status

		if lastStatus != analysis.StatusCompleted {
			// Error caused by an issue with the package (probably).
			// Don't continue with phases if this one did not complete successfully.
			break
		}
	}

	if lastError != nil {
		LogDynamicAnalysisError(pkg, lastRunPhase, lastError)
	} else {
		LogDynamicAnalysisResult(pkg, lastRunPhase, lastStatus)
	}

	return results, lastRunPhase, lastStatus, lastError
}
