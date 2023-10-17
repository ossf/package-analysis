package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"github.com/ossf/package-analysis/internal/featureflags"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

// sandboxExecutionLogPath is the absolute path of the execution log file
// inside the sandbox. The file is used for code execution feature.
const sandboxExecutionLogPath = "/execution.log"

var nonSpaceControlChars = regexp.MustCompile("[\x00-\x08\x0b-\x1f\x7f]")

func isExecutePhaseEnabled(ecosystem pkgecosystem.Ecosystem) bool {
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

// initialiseExecutePhase copies an empty file to the sandbox, which becomes the execution log.
func initialiseExecutePhase(ctx context.Context, sb sandbox.Sandbox) error {
	tempFile, err := os.CreateTemp("", "")
	if err != nil {
		return fmt.Errorf("could not create empty execution log file in host: %w", err)
	}

	// file wasn't written to, so don't worry too much about close errors
	_ = tempFile.Close()
	tempPath := tempFile.Name()

	if err := sb.CopyIntoSandbox(ctx, tempPath, sandboxExecutionLogPath); err != nil {
		return fmt.Errorf("could not copy empty execution log file to sandbox: %w", err)
	}

	// file wasn't written to, so don't worry too much about remove errors
	_ = os.Remove(tempPath)

	return nil
}

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
