package worker

import (
	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/dynamicanalysis"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/sandbox"
)

type DynamicAnalysisResults map[pkgecosystem.RunPhase]*dynamicanalysis.Result

/*
RunDynamicAnalysis runs dynamic analysis on the given package in the sandbox
provided, across all phases (e.g. import, install) valid in the package ecosystem.
Status and errors are logged to stdout. There are 3 return values:

results: Map of each successfully run phase to the corresponding dynamic analysis result.
If error is not nil, then results[lastRunPhase] is nil.

lastRunPhase: the last phase that was run. If error is non-nil, this phase did not complete successfully
and the results for this phase are not recorded. Otherwise, results[lastRunPhase] contains
the corresponding results this phase, including any abnormal termination of the sandboxed process.

error: Any error that occurred in the runtime/sandbox infrastructure. This does not include errors caused
by the package under analysis.
*/
func RunDynamicAnalysis(sb sandbox.Sandbox, pkg *pkgecosystem.Pkg) (results DynamicAnalysisResults, lastRunPhase pkgecosystem.RunPhase, err error) {
	results = make(DynamicAnalysisResults)
	for _, phase := range pkg.Manager().RunPhases() {
		result, err := dynamicanalysis.Run(sb, pkg.Command(phase))
		lastRunPhase = phase

		if err != nil {
			// Error when trying to actually run; don't record the result for this phase
			// or attempt subsequent phases
			break
		}

		results[phase] = result

		if result.Status != analysis.StatusCompleted {
			// Error caused by an issue with the package (probably).
			// Don't continue with phases if this one did not complete successfully.
			break
		}
	}

	if err != nil {
		LogDynamicAnalysisError(pkg, lastRunPhase, err)
	} else {
		// Produce a log message for the final status to help generate metrics.
		LogDynamicAnalysisResult(pkg, lastRunPhase, results[lastRunPhase].Status)
	}

	return results, lastRunPhase, err
}
