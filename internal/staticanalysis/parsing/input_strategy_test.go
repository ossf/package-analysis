package parsing

import (
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func TestMultipleFileInput(t *testing.T) {
	tests := []struct {
		name      string
		filePaths []string
		wantErr   bool
	}{
		{
			name:      "simple",
			filePaths: []string{"test1.txt", "test2.txt"},
			wantErr:   false,
		},
	}

	// white box test of MultipleFileInput.SendTo()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workingDir := t.TempDir()
			cmd := exec.Command("")
			input := MultipleFileInput(tt.filePaths)
			err := input.SendTo(cmd, workingDir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error or absence of error: error = %v", err)
			} else if tt.wantErr {
				return
			}
			inputFile := workingDir + string(os.PathSeparator) + "input.txt"
			expectedArgs := []string{"--batch", inputFile}
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
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "simple",
			filePath: "test1.txt",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("")
			input := SingleFileInput(tt.filePath)
			err := input.SendTo(cmd, "")
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error or absence of error: error = %v", err)
			} else if tt.wantErr {
				return
			}
			expectedArgs := []string{"--file", "test1.txt"}
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
		wantErr     bool
	}{
		{
			name:        "simple",
			inputString: "abcdefghijklmnopqrtsuvwxyz",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("")
			input := StringInput(tt.inputString)
			if err := input.SendTo(cmd, ""); (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error or absence of error: error = %v", err)
			} else if tt.wantErr {
				return
			}
			expectedArgs := []string{}
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
