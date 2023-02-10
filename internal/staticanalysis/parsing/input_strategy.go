package parsing

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

/*
InputStrategy allows different ways of passing input to a parser, for example
a list of files or a raw string. Each instance represents specific input data
*/
type InputStrategy interface {
	/*
		SendTo sends the input data held by this object to the given command. This
		may involve either adding arguments or redirecting stdin, as well as file IO.
		tempDir is a path to a temporary directory that may be used by the strategy
		to write intermediate files. It is the caller's responsibility to clean up
		the directory when the parser has been run.
	*/
	SendTo(parserCmd *exec.Cmd, tempDir string) error
}

type stringInput struct {
	rawString string
}

type singleFileInput struct {
	filePath string
}

type multipleFileInput struct {
	filePaths []string
}

func StringInput(sourceCode string) InputStrategy {
	return stringInput{rawString: sourceCode}
}

func SingleFileInput(path string) InputStrategy {
	return singleFileInput{filePath: path}
}

func MultipleFileInput(paths []string) InputStrategy {
	return multipleFileInput{filePaths: paths}
}

func (s stringInput) SendTo(parserCmd *exec.Cmd, tempDir string) error {
	// create a pipe to send the source code to the parser via stdin
	pipe, pipeErr := parserCmd.StdinPipe()
	if pipeErr != nil {
		return fmt.Errorf("failed to create pipe: %v", pipeErr)
	}

	if _, pipeErr = pipe.Write([]byte(s.rawString)); pipeErr != nil {
		return fmt.Errorf("failed to write source string to pipe: %w", pipeErr)
	}

	if pipeErr = pipe.Close(); pipeErr != nil {
		return fmt.Errorf("failed to close pipe: %w", pipeErr)
	}

	return nil
}

func (s singleFileInput) SendTo(parserCmd *exec.Cmd, tempDir string) error {
	parserCmd.Args = append(parserCmd.Args, "--file", s.filePath)
	return nil
}

func (m multipleFileInput) SendTo(parserCmd *exec.Cmd, tempDir string) error {
	// write input file paths to temp file
	infilePath := tempDir + string(os.PathSeparator) + "input.txt"
	parserCmd.Args = append(parserCmd.Args, "--batch", infilePath)

	filePathData := []byte(strings.Join(m.filePaths, "\n"))
	if err := os.WriteFile(infilePath, filePathData, 0o666); err != nil {
		return fmt.Errorf("runParser failed to write file paths to temp file: %w", err)
	}
	return nil
}
