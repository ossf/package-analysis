package analysis

import (
	"github.com/ossf/package-analysis/internal/dynamicanalysis"
	"github.com/ossf/package-analysis/internal/log"
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
	for _, phase := range pkg.Manager().RunPhases() {
		result, err := dynamicanalysis.Run(sb, pkg.Command(phase))
		lastRunPhase = phase

		if err != nil {
			// Error when trying to actually run; don't record the result for this phase
			// or attempt subsequent phases
			break
		}

		results[phase] = result

		if result.Status != dynamicanalysis.StatusCompleted {
			// Error caused by an issue with the package (probably).
			// Don't continue with phases if this one did not complete successfully.
			break
		}
	}

	return results, lastRunPhase, err
}

// LogDynamicAnalysisError indicates some error happened while attempting to run
// the package code, which was not caused by the package itself. This means it was
// not possible to analyse the package properly, and the results are invalid.
func LogDynamicAnalysisError(pkg *pkgecosystem.Pkg, errorPhase pkgecosystem.RunPhase, err error) {
	log.Error("Analysis run failed",
		log.Label("ecosystem", pkg.Ecosystem()),
		log.Label("name", pkg.Name()),
		log.Label("phase", string(errorPhase)),
		log.Label("version", pkg.Version()),
		"error", err)
}

// LogDynamicAnalysisResult indicates that the package code was run successfully,
// and what happened when it was run. This may include errors in the analysis
// of the package, but not errors in the running itself.
func LogDynamicAnalysisResult(pkg *pkgecosystem.Pkg, finalPhase pkgecosystem.RunPhase, finalStatus dynamicanalysis.Status) {
	ecosystem := pkg.Ecosystem()
	name := pkg.Name()
	version := pkg.Version()
	lastPhase := string(finalPhase)

	switch finalStatus {
	case dynamicanalysis.StatusCompleted:
		log.Info("Analysis completed sucessfully",
			log.Label("ecosystem", ecosystem),
			log.Label("name", name),
			log.Label("version", version),
			log.Label("last_phase", lastPhase))

	case dynamicanalysis.StatusErrorAnalysis:
		log.Warn("Analysis error - analysis",
			log.Label("ecosystem", ecosystem),
			log.Label("name", name),
			log.Label("version", version),
			log.Label("last_phase", lastPhase))
	case dynamicanalysis.StatusErrorTimeout:
		log.Warn("Analysis error - timeout",
			log.Label("ecosystem", ecosystem),
			log.Label("name", name),
			log.Label("version", version),
			log.Label("last_phase", lastPhase))
	case dynamicanalysis.StatusErrorOther:
		log.Warn("Analysis error - other",
			log.Label("ecosystem", ecosystem),
			log.Label("name", name),
			log.Label("version", version),
			log.Label("last_phase", lastPhase))
	}
}
