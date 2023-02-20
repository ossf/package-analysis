package externalcmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

/*
Input allows different ways of passing input to an external command, for example
a list of files or a raw string. Each instance represents specific input data.
*/
type Input interface {
	/*
		SendTo sends the input data held by this object to the given command. This
		may involve either adding arguments or redirecting stdin, as well as file IO.
		tempDir is a path to a temporary directory that may be used by the strategy
		to write intermediate files. It is the caller's responsibility to clean up
		the directory when the parser has been run.
	*/
	SendTo(cmd *exec.Cmd, argHandler InputArgHandler, tempDir string) error
}

/*
InputArgHandler abstracts command-specific behaviour for how to pass files
as command line arguments, including which option strings to use.
*/
type InputArgHandler interface {
	// ReadStdinArg returns the command line arguments which
	// specify that the command should read input from stdin
	ReadStdinArg() []string

	// SingleFileArg returns the command line arguments which
	// specify the path to a single input file for the command.
	SingleFileArg(filePath string) []string

	// FileListArg returns the command line arguments which
	// specify the path to a file that contains a list of
	// files that should be processed by the command
	FileListArg(fileListPath string) []string
}

type stringInput struct {
	input string
}

type singleFileInput struct {
	filePath string
}

type multipleFileInput struct {
	filePaths []string
}

func StringInput(rawInput string) Input {
	return stringInput{input: rawInput}
}

func SingleFileInput(path string) Input {
	return singleFileInput{filePath: path}
}

func MultipleFileInput(paths []string) Input {
	return multipleFileInput{filePaths: paths}
}

func (s stringInput) SendTo(cmd *exec.Cmd, argHandler InputArgHandler, tempDir string) error {
	cmd.Args = append(cmd.Args, argHandler.SingleFileArg("-")...)

	// create a pipe to send the source code to the parser via stdin
	pipe, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %w", err)
	}

	if _, err := pipe.Write([]byte(s.input)); err != nil {
		return fmt.Errorf("failed to write source string to pipe: %w", err)
	}

	if err := pipe.Close(); err != nil {
		return fmt.Errorf("failed to close pipe: %w", err)
	}

	return nil
}

func (s singleFileInput) SendTo(cmd *exec.Cmd, args InputArgHandler, tempDir string) error {
	cmd.Args = append(cmd.Args, args.SingleFileArg(s.filePath)...)
	return nil
}

func (m multipleFileInput) SendTo(cmd *exec.Cmd, argHandler InputArgHandler, tempDir string) error {
	// write input file paths to temp file
	infilePath := filepath.Join(tempDir, "input.txt")
	cmd.Args = append(cmd.Args, argHandler.FileListArg(infilePath)...)

	filePathData := []byte(strings.Join(m.filePaths, "\n"))
	if err := os.WriteFile(infilePath, filePathData, 0o666); err != nil {
		return fmt.Errorf("runParser failed to write file paths to temp file: %w", err)
	}
	return nil
}
