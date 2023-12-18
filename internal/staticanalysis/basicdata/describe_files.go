package basicdata

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/externalcmd"
)

// fileCmdInputArgs describes how to pass file arguments to the `file` command.
type fileCmdArgsHandler struct{}

func (h fileCmdArgsHandler) SingleFileArg(filePath string) []string {
	return []string{filePath}
}

func (h fileCmdArgsHandler) FileListArg(fileListPath string) []string {
	return []string{"--files-from", fileListPath}
}

func (h fileCmdArgsHandler) ReadStdinArg() []string {
	// reads file list from standard input
	return h.FileListArg("-")
}

func detectFileTypes(ctx context.Context, paths []string) ([]string, error) {
	workingDir, err := os.MkdirTemp("", "package-analysis-basic-data-*")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(workingDir); err != nil {
			slog.ErrorContext(ctx, "could not remove working directory", "path", workingDir, "error", err)
		}
	}()

	cmd := exec.CommandContext(ctx, "file", "--brief")
	input := externalcmd.MultipleFileInput(paths)

	if err := input.SendTo(cmd, fileCmdArgsHandler{}, workingDir); err != nil {
		return nil, fmt.Errorf("failed to prepare input: %w", err)
	}

	fileCmdOutput, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running file command: %w", err)
	}

	descriptionsString := strings.TrimSpace(string(fileCmdOutput))
	if descriptionsString == "" {
		// no files input, probably
		return []string{}, nil
	}

	// command output is newline-separated list of file types,
	// with the order matching the input file list.
	return strings.Split(descriptionsString, "\n"), nil
}
