package analysis

import (
	"github.com/ossf/package-analysis/internal/dynamicanalysis"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/sandbox"
)

type DynamicAnalysisResults map[pkgecosystem.RunPhase]*dynamicanalysis.Result

// RunDynamicAnalysis
// Runs dynamic analysis on the given package in the sandbox provided, and returns the results and error (if any)
// at all registered running phases for the package ecosystem. Status is logged to standard output if requested.
// The return value maps each successfully run phase to the corresponding dynamic analysis result.
func RunDynamicAnalysis(sb sandbox.Sandbox, pkg *pkgecosystem.Pkg, logStatus bool) (results DynamicAnalysisResults, err error) {
	var lastNonErrorPhase pkgecosystem.RunPhase
	var lastNonErrorResult *dynamicanalysis.Result

	for _, phase := range phasesToAnalyze(pkg) {
		result, err := dynamicanalysis.Run(sb, pkg.Command(phase))
		if err != nil {
			// Error when trying to actually run; don't record the result for this phase
			// or attempt subsequent phases
			if logStatus {
				logDynamicAnalysisError(pkg, phase, err)
			}
			return results, err
		}

		results[phase] = result
		lastNonErrorPhase = phase
		lastNonErrorResult = result

		if result.Status != dynamicanalysis.StatusCompleted {
			// Error caused by an issue with the package (probably).
			// Don't continue processing subsequent phases if this one did not complete successfully.
			break
		}
	}

	if logStatus {
		// Produce a log message for the final status to help generate metrics.
		logDynamicAnalysisResult(pkg, lastNonErrorPhase, lastNonErrorResult)
	}

	return results, nil
}

// logDynamicAnalysisError
// Indicates some error happened while attempting to run the package code, which was not caused by the package itself.
// This kind of error means it was not possible to analyse the package properly, and the results are invalid.
func logDynamicAnalysisError(pkg *pkgecosystem.Pkg, errorPhase pkgecosystem.RunPhase, err error) {
	log.Error("Analysis run failed",
		log.Label("ecosystem", pkg.Ecosystem()),
		log.Label("name", pkg.Name()),
		log.Label("phase", string(errorPhase)),
		log.Label("version", pkg.Version()),
		"error", err)
}

// logDynamicAnalysisResult
// Indicates that the package code was run successfully, and what happened when it was run
// This may include errors in the analysis of the package, but not errors in the running itself.
func logDynamicAnalysisResult(pkg *pkgecosystem.Pkg, finalPhase pkgecosystem.RunPhase, finalResult *dynamicanalysis.Result) {
	ecosystem := pkg.Ecosystem()
	name := pkg.Name()
	version := pkg.Version()
	lastPhase := string(finalPhase)
	finalStatus := finalResult.Status

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

// phasesToAnalyze
// Logic to decide which phases of dynamic analysis to run
// Currently just returns all possible phases for the package ecosystem
func phasesToAnalyze(pkg *pkgecosystem.Pkg) []pkgecosystem.RunPhase {
	return pkg.Manager().RunPhases()
}
