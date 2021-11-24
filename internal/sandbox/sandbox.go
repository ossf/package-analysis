package sandbox

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
)

const (
	podmanBin  = "podman"
	runtimeBin = "/usr/bin/runsc"
	rootDir    = "/var/run/runsc"
	logPath    = "/tmp/runsc.log.boot"
)

type RunStatus uint8

const (
	// RunStatusUnknown is used when some other issue occured that prevented
	// an attempt to run the command.
	RunStatusUnknown = iota

	// RunStatusSuccess is used to indicate that the command being executed
	// successfully.
	RunStatusSuccess

	// RunStatusFailure is used to indicate that the command exited with some
	// failure.
	RunStatusFailure

	// RunStatusTimeout is used to indicate that the command failed to complete
	// within the allowed timeout.
	RunStatusTimeout
)

type RunResult struct {
	Status RunStatus
	Stderr io.Reader
	Stdout io.Reader
}

// Log returns the log file recorded during a run.
//
// This log will contain strace data.
func (r *RunResult) Log() (io.ReadCloser, error) {
	return os.Open(logPath)
}

func podmanPull(image string) error {
	args := []string{"pull", image}
	cmd := exec.Command(podmanBin, args...)
	return cmd.Run()
}

func podmanPrune() error {
	args := []string{"image", "prune", "-f"}
	cmd := exec.Command(podmanBin, args...)
	return cmd.Run()
}

func podmanCleanContainers() error {
	args := []string{"rm", "--all", "--force"}
	cmd := exec.Command(podmanBin, args...)
	return cmd.Run()
}

func podmanRunCmd(image, command string, extraArgs []string) *exec.Cmd {
	args := []string{
		"run",
		"--runtime=" + runtimeBin,
		"--runtime-flag=root=" + rootDir,
		"--runtime-flag=debug-log=/tmp/runsc.log.%COMMAND%",
		"--runtime-flag=net-raw",
		"--runtime-flag=debug",
		"--runtime-flag=strace",
		"--runtime-flag=log-packets",
		"--cgroup-manager=cgroupfs",
		"--events-backend=file",
		"--hostname=box",
		"--rm",
	}
	args = append(args, extraArgs...)
	args = append(args, image, "sh", "-c", command)

	cmd := exec.Command(podmanBin, args...)
	return cmd
}

type Sandbox struct {
	image string
}

// Initializes the Sandbox ready for running commands.
//
// The image supplied will be pulled if it hasn't already been.
func Init(image string) (*Sandbox, error) {
	if err := podmanPull(image); err != nil {
		return nil, err
	}
	if err := podmanPrune(); err != nil {
		return nil, err
	}
	return &Sandbox{
		image: image,
	}, nil
}

// Run will run a single command inside the sandbox.
//
// The container used to execute the command will be removed when the command
// is completed.
//
// This function is useful for running multiple commands in the sandbox.
func (s *Sandbox) Run(command string, args ...string) (*RunResult, error) {
	// Delete existing logs (if any).
	// This function uses a fixed log name and is not threadsafe.
	if err := os.RemoveAll(logPath); err != nil {
		return &RunResult{}, err
	}

	cmd := podmanRunCmd(s.image, command, args)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return &RunResult{}, err
	}

	// Wire up the run result.
	result := &RunResult{
		Status: RunStatusSuccess,
		Stdout: &stdout,
		Stderr: &stderr,
	}

	err := cmd.Wait()
	if err != nil {
		// Ignore the error if stderr contains "gofer is still running"
		if !strings.Contains(stderr.String(), "gofer is still running") {
			result.Status = RunStatusFailure
		}
	}

	return result, err
}

// Clean stops and removes all containers.
func (s *Sandbox) Clean() error {
	// remove log files too?
	return podmanCleanContainers()
}
