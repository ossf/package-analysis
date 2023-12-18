package worker

import (
	"context"
	"errors"
	"log/slog"
	"os/exec"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
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
)

// LogDynamicAnalysisError indicates some error happened while attempting to run
// the package code, which was not caused by the package itself. This means it was
// not possible to analyse the package properly, and the results are invalid.
func LogDynamicAnalysisError(ctx context.Context, pkg *pkgmanager.Pkg, errorPhase analysisrun.DynamicPhase, err error) {
	slog.ErrorContext(ctx, runErrorLogMsg,
		log.Label("phase", string(errorPhase)),
		"error", err)

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		slog.DebugContext(ctx, "Command stderr", "stderr", exitErr.Stderr)
	}
}

// LogDynamicAnalysisResult indicates that the package code was run successfully,
// and what happened when it was run. This may include errors in the analysis
// of the package, but not errors in the running itself.
func LogDynamicAnalysisResult(ctx context.Context, pkg *pkgmanager.Pkg, finalPhase analysisrun.DynamicPhase, finalStatus analysis.Status) {
	labels := []interface{}{
		log.Label("last_phase", string(finalPhase)),
	}

	switch finalStatus {
	case analysis.StatusCompleted:
		slog.InfoContext(ctx, analysisCompleteLogMsg, labels...)
	case analysis.StatusErrorAnalysis:
		slog.WarnContext(ctx, analysisErrorLogMsg, labels...)
	case analysis.StatusErrorTimeout:
		slog.WarnContext(ctx, timeoutErrorLogMsg, labels...)
	case analysis.StatusErrorOther:
		slog.WarnContext(ctx, otherErrorLogMsg, labels...)
	}
}
