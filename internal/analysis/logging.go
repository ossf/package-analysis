package analysis

import (
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
)

/*
NOTE: These strings are referenced externally by infrastructure for dashboard
reporting / metrics purposes, and so should be changed with care.

See file infra/terraform/metrics/log_metrics.tf
*/
const analysisCompleteLogMsg = "Analysis completed sucessfully" // TODO sucessfully -> successfully
const analysisErrorLogMsg = "Analysis error - analysis"
const timeoutErrorLogMsg = "Analysis error - timeout"
const otherErrorLogMsg = "Analysis error - other"
const runErrorLogMsg = "Analysis run failed"
const GotRequestLogMsg = "Got request"

// LogDynamicAnalysisError indicates some error happened while attempting to run
// the package code, which was not caused by the package itself. This means it was
// not possible to analyse the package properly, and the results are invalid.
func LogDynamicAnalysisError(pkg *pkgecosystem.Pkg, errorPhase pkgecosystem.RunPhase, err error) {
	log.Error(runErrorLogMsg,
		log.Label("ecosystem", pkg.Ecosystem()),
		log.Label("name", pkg.Name()),
		log.Label("phase", string(errorPhase)),
		log.Label("version", pkg.Version()),
		"error", err)
}

// LogDynamicAnalysisResult indicates that the package code was run successfully,
// and what happened when it was run. This may include errors in the analysis
// of the package, but not errors in the running itself.
func LogDynamicAnalysisResult(pkg *pkgecosystem.Pkg, finalPhase pkgecosystem.RunPhase, finalStatus Status) {
	ecosystem := pkg.Ecosystem()
	name := pkg.Name()
	version := pkg.Version()
	lastPhase := string(finalPhase)

	labels := []interface{}{
		log.Label("ecosystem", ecosystem),
		log.Label("name", name),
		log.Label("version", version),
		log.Label("last_phase", lastPhase),
	}

	switch finalStatus {
	case StatusCompleted:
		log.Info(analysisCompleteLogMsg, labels...)
	case StatusErrorAnalysis:
		log.Warn(analysisErrorLogMsg, labels...)
	case StatusErrorTimeout:
		log.Warn(timeoutErrorLogMsg, labels...)
	case StatusErrorOther:
		log.Warn(otherErrorLogMsg, labels...)
	}
}
