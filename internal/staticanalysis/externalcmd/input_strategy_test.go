package externalcmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

// optionArgHandler is used to test option/value type argument passing.
type optionArgHandler struct{}

func (a optionArgHandler) ReadStdinArg() []string {
	return []string{"--read-stdin"}
}

func (a optionArgHandler) SingleFileArg(filePath string) []string {
	return []string{"--single-file", filePath}
}

func (a optionArgHandler) FileListArg(fileListPath string) []string {
	return []string{"--file-list", fileListPath}
}

// positionalArgHandler is used to test positional (or no) arg passing.
type positionalArgHandler struct{}

func (a positionalArgHandler) ReadStdinArg() []string {
	return []string{} // no args
}

func (a positionalArgHandler) SingleFileArg(filePath string) []string {
	return []string{filePath}
}

func (a positionalArgHandler) FileListArg(fileListPath string) []string {
	return []string{fileListPath}
}

func TestMultipleFileInput(t *testing.T) {
	tests := []struct {
		name       string
		filePaths  []string
		argHandler InputArgHandler
		wantErr    bool
	}{
		{
			name:       "positional args",
			filePaths:  []string{"test1.txt", "test2.txt"},
			argHandler: positionalArgHandler{},
			wantErr:    false,
		},
		{
			name:       "verbose args",
			filePaths:  []string{"test1.txt", "test2.txt"},
			argHandler: optionArgHandler{},
			wantErr:    false,
		},
	}

	// white box test of MultipleFileInput.SendTo()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workingDir := t.TempDir()
			cmd := exec.Command("")
			input := MultipleFileInput(tt.filePaths)
			err := input.SendTo(cmd, tt.argHandler, workingDir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error or absence of error: error = %v", err)
			} else if tt.wantErr {
				return
			}
			inputFile := filepath.Join(workingDir, "input.txt")
			expectedArgs := tt.argHandler.FileListArg(inputFile)
			if reflect.DeepEqual(expectedArgs, cmd.Args) {
				t.Errorf("expected cmd args did not match actual. Expected: %v, actual: %v", expectedArgs, cmd.Args)
			}
			filePaths, err := os.ReadFile(inputFile)
			expectedContents := "test1.txt\ntest2.txt"
			if string(filePaths) != expectedContents {
				t.Errorf("expected input file contents did not match actual. Expected: %v, actual: %v", expectedContents, filePaths)
			}
		})
	}
}

func TestSingleFileInput(t *testing.T) {
	tests := []struct {
		name       string
		filePath   string
		argHandler InputArgHandler
		wantErr    bool
	}{
		{
			name:       "verbose args",
			filePath:   "test1.txt",
			argHandler: optionArgHandler{},
			wantErr:    false,
		},
		{
			name:       "positional args",
			filePath:   "test1.txt",
			argHandler: positionalArgHandler{},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("")
			input := SingleFileInput(tt.filePath)
			err := input.SendTo(cmd, tt.argHandler, "")
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error or absence of error: error = %v", err)
			} else if tt.wantErr {
				return
			}
			expectedArgs := tt.argHandler.SingleFileArg(tt.filePath)
			if reflect.DeepEqual(expectedArgs, cmd.Args) {
				t.Errorf("expected cmd args did not match actual. Expected: %v, actual: %v", expectedArgs, cmd.Args)
			}
		})
	}
}

func TestStringInput(t *testing.T) {
	tests := []struct {
		name        string
		inputString string
		argHandler  InputArgHandler
		wantErr     bool
	}{
		{
			name:        "verbose args",
			inputString: "abcdefghijklmnopqrtsuvwxyz",
			argHandler:  optionArgHandler{},
			wantErr:     false,
		},
		{
			name:        "positional args",
			inputString: "abcdefghijklmnopqrtsuvwxyz",
			argHandler:  positionalArgHandler{},
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("")
			input := StringInput(tt.inputString)
			if err := input.SendTo(cmd, tt.argHandler, ""); (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error or absence of error: error = %v", err)
			} else if tt.wantErr {
				return
			}
			expectedArgs := tt.argHandler.ReadStdinArg()
			if reflect.DeepEqual(expectedArgs, cmd.Args) {
				t.Errorf("expected cmd args did not match actual. Expected: %v, actual: %v", expectedArgs, cmd.Args)
			}
			buffer := make([]byte, len(tt.inputString))
			_, err := cmd.Stdin.Read(buffer)
			if err != nil {
				t.Fatalf("cmd.Stdin read error: %v", err)
			}
			if reflect.DeepEqual(tt.inputString, buffer) {
				t.Fatalf("pipe did not contain expected input. Expected: %s, actual: %v", tt.inputString, buffer)
			}
		})
	}
}
