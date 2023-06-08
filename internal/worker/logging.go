package worker

import (
	"errors"
	"os/exec"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

/*
NOTE: These strings are referenced externally by infrastructure for dashboard
reporting / metrics purposes, and so should be changed with care.

See file infra/terraform/metrics/log_metrics.tf.
*/
const (
	analysisCompleteLogMsg = "Analysis completed sucessfully" // TODO sucessfully -> successfully
	analysisErrorLogMsg    = "Analysis error - analysis"
	timeoutErrorLogMsg     = "Analysis error - timeout"
	otherErrorLogMsg       = "Analysis error - other"
	runErrorLogMsg         = "Analysis run failed"
	gotRequestLogMsg       = "Got request"
)

// LogDynamicAnalysisError indicates some error happened while attempting to run
// the package code, which was not caused by the package itself. This means it was
// not possible to analyse the package properly, and the results are invalid.
func LogDynamicAnalysisError(pkg *pkgmanager.Pkg, errorPhase analysisrun.DynamicPhase, err error) {
	log.Error(runErrorLogMsg,
		log.Label("ecosystem", pkg.EcosystemName()),
		log.Label("phase", string(errorPhase)),
		"name", pkg.Name(),
		"version", pkg.Version(),
		"error", err)

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		log.Debug("Command stderr", "stderr", exitErr.Stderr)
	}
}

// LogDynamicAnalysisResult indicates that the package code was run successfully,
// and what happened when it was run. This may include errors in the analysis
// of the package, but not errors in the running itself.
func LogDynamicAnalysisResult(pkg *pkgmanager.Pkg, finalPhase analysisrun.DynamicPhase, finalStatus analysis.Status) {
	labels := []interface{}{
		log.Label("ecosystem", pkg.EcosystemName()),
		log.Label("last_phase", string(finalPhase)),
		"name", pkg.Name(),
		"version", pkg.Version(),
	}

	switch finalStatus {
	case analysis.StatusCompleted:
		log.Info(analysisCompleteLogMsg, labels...)
	case analysis.StatusErrorAnalysis:
		log.Warn(analysisErrorLogMsg, labels...)
	case analysis.StatusErrorTimeout:
		log.Warn(timeoutErrorLogMsg, labels...)
	case analysis.StatusErrorOther:
		log.Warn(otherErrorLogMsg, labels...)
	}
}

// LogRequest records that a request for analysis was received by the worker.
func LogRequest(e pkgecosystem.Ecosystem, name, version, localPath, resultsBucketOverride string) {
	log.Info(gotRequestLogMsg,
		log.Label("ecosystem", e.String()),
		"name", name,
		"version", version,
		"package_path", localPath,
		"results_bucket_override", resultsBucketOverride,
	)
}
