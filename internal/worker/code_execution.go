package worker

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"github.com/ossf/package-analysis/internal/sandbox"
)

// sandboxExecutionLogPath is the absolute path of the execution log file
// inside the sandbox. This file is used for logging during the execute phase.
const sandboxExecutionLogPath = "/execution.log"

var nonSpaceControlChars = regexp.MustCompile("[\x00-\x08\x0b-\x1f\x7f]")

// retrieveExecutionLog copies the execution log back from the sandbox
// to the host, so it can be included in the dynamic analysis results.
// To mitigate against binary code injection, all control characters except
// tab and newline are stripped from the file.
func retrieveExecutionLog(ctx context.Context, sb sandbox.Sandbox) (string, error) {
	executionLogDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	defer os.RemoveAll(executionLogDir)
	hostExecutionLogPath := filepath.Join(executionLogDir, "execution.log")

	// if the copy fails, it could be that the execution log is not actually present.
	// For now, we'll just log the error and otherwise ignore it
	if err := sb.CopyBackToHost(ctx, hostExecutionLogPath, sandboxExecutionLogPath); err != nil {
		slog.WarnContext(ctx, "Could not retrieve execution log from sandbox", "error", err)
		return "", nil
	}

	logData, err := os.ReadFile(hostExecutionLogPath)
	if err != nil {
		return "", err
	}

	// remove control characters except tab (\x09) and newline (\x0A)
	processedLog := nonSpaceControlChars.ReplaceAllLiteral(logData, []byte{})
	slog.InfoContext(ctx, "Read execution log", "rawLength", len(logData), "processedLength", len(processedLog))

	return string(processedLog), nil
}
