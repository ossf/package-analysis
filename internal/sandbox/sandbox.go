package sandbox

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/ossf/package-analysis/internal/log"
)

const (
	podmanBin       = "podman"
	runtimeBin      = "/usr/local/bin/runsc_compat.sh"
	rootDir         = "/var/run/runsc"
	straceFile      = "runsc.log.boot"
	hostname        = "box"
	containerPrefix = "box"
	logDirPattern   = "sandbox_logs_"
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
	logPath string
	Status  RunStatus
	Stderr  io.Reader
	Stdout  io.Reader
}

// Log returns the log file recorded during a run.
//
// This log will contain strace data.
func (r *RunResult) Log() (io.ReadCloser, error) {
	return os.Open(r.logPath)
}

type Sandbox interface {
	// Run will run the sandbox for the given args.
	//
	// The container used to execute the command will be reused until Clean()
	// is called.
	Run(...string) (*RunResult, error)

	// Clean cleans up a Sandbox.
	//
	// Once called the Sandbox cannot be used again.
	Clean() error
}

// volume represents a volume mapping between a host src and a container dest
type volume struct {
	src  string
	dest string
}

func (v volume) args() []string {
	return []string{
		"-v",
		fmt.Sprintf("%s:%s", v.src, v.dest),
	}
}

// Implements the Sandbox interface using "podman".
type podmanSandbox struct {
	image     string
	tag       string
	id        string
	container string
	noPull    bool
	volumes   []volume
}

type Option interface{ set(*podmanSandbox) }
type option func(*podmanSandbox)       // option implements Option.
func (o option) set(sb *podmanSandbox) { o(sb) }

func New(image string, options ...Option) Sandbox {
	sb := &podmanSandbox{
		image:     image,
		tag:       "",
		container: "",
		noPull:    false,
		volumes:   make([]volume, 0),
	}
	for _, o := range options {
		o.set(sb)
	}
	return sb
}

// NoPull will disable the image for the sandbox from being pulled during Init.
func NoPull() Option {
	return option(func(sb *podmanSandbox) { sb.noPull = true })
}

// Volume can be used to specify an additional volume map into the container.
//
// src is the path in the host that will be mapped to the dest path.
func Volume(src, dest string) Option {
	return option(func(sb *podmanSandbox) {
		sb.volumes = append(sb.volumes, volume{
			src:  src,
			dest: dest,
		})
	})
}

func Tag(tag string) Option {
	return option(func(sb *podmanSandbox) { sb.tag = tag })
}

func removeAllLogs() error {
	matches, err := filepath.Glob(path.Join(os.TempDir(), logDirPattern+"*"))
	if err != nil {
		return err
	}
	for _, m := range matches {
		if err := os.RemoveAll(m); err != nil {
			return err
		}
	}
	return nil
}

func podman(args ...string) *exec.Cmd {
	args = append([]string{
		"--cgroup-manager=cgroupfs",
		"--events-backend=file",
	}, args...)
	log.Debug("podman", "args", args)
	return exec.Command(podmanBin, args...)
}

func podmanRun(args ...string) error {
	cmd := podman(args...)
	return cmd.Run()
}

func podmanPrune() error {
	return podmanRun("image", "prune", "-f")
}

func podmanCleanContainers() error {
	return podmanRun("rm", "--all", "--force")
}

func (s *podmanSandbox) pullImage() error {
	return podmanRun("pull", s.imageWithTag())
}

func (s *podmanSandbox) createContainer() (string, error) {
	args := []string{
		"create",
		"--runtime=" + runtimeBin,
		"--init",
		"--hostname=" + hostname,
		s.imageWithTag(),
	}
	args = append(args, s.extraArgs()...)
	cmd := podman(args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(buf.Bytes())), nil
}

func (s *podmanSandbox) startContainerCmd(logDir string) *exec.Cmd {
	return podman(
		"start",
		"--runtime="+runtimeBin,
		"--runtime-flag=root="+rootDir,
		"--runtime-flag=debug-log="+path.Join(logDir, "runsc.log.%COMMAND%"),
		"--runtime-flag=net-raw",
		"--runtime-flag=strace",
		"--runtime-flag=log-packets",
		s.container)
}

func (s *podmanSandbox) stopContainerCmd() *exec.Cmd {
	return podman("stop", s.container)
}

func (s *podmanSandbox) forceStopContainer() error {
	return podmanRun(
		"stop",
		"-t=5", // Wait a max of 5 seconds for a graceful stop.
		"-i",   // Ignore any errors of the specified container not being in the store.
		s.container)
}

func (s *podmanSandbox) execContainerCmd(execArgs []string) *exec.Cmd {
	args := []string{
		"exec",
		s.container,
	}
	args = append(args, execArgs...)
	return podman(args...)
}

func (s *podmanSandbox) extraArgs() []string {
	args := make([]string, 0)
	for _, v := range s.volumes {
		args = append(args, v.args()...)
	}
	return args
}

func (s *podmanSandbox) imageWithTag() string {
	tag := "latest"
	if s.tag != "" {
		tag = s.tag
	}
	return fmt.Sprintf("%s:%s", s.image, tag)
}

// init initializes the sandbox.
func (s *podmanSandbox) init() error {
	if s.container != "" {
		return nil
	}
	// Delete existing logs (if any).
	if err := removeAllLogs(); err != nil {
		return err
	}
	if !s.noPull {
		if err := s.pullImage(); err != nil {
			return err
		}
	}
	if err := podmanPrune(); err != nil {
		return err
	}
	if id, err := s.createContainer(); err == nil {
		s.container = id
	} else {
		return err
	}
	return nil
}

// Run will run a single command inside the sandbox.
//
// The container used to execute the command will be removed when the command
// is completed.
//
// This function is useful for running multiple commands in the sandbox.
func (s *podmanSandbox) Run(args ...string) (*RunResult, error) {
	if err := s.init(); err != nil {
		return &RunResult{}, err
	}

	// Create a place to stash the logs for this run.
	logDir, err := os.MkdirTemp("", logDirPattern)
	if err != nil {
		return &RunResult{}, err
	}

	// Prepare the run result.
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	result := &RunResult{
		logPath: path.Join(logDir, straceFile),
		Status:  RunStatusUnknown,
		Stdout:  &stdout,
		Stderr:  &stderr,
	}

	// Prepare stdout and stderr writers
	logOut := log.Writer(log.InfoLevel,
		"args", args)
	defer logOut.Close()
	logErr := log.Writer(log.WarnLevel,
		"args", args)
	defer logErr.Close()
	outWriter := io.MultiWriter(&stdout, logOut)
	errWriter := io.MultiWriter(&stderr, logErr)

	// Start the container
	startCmd := s.startContainerCmd(logDir)
	startCmd.Stdout = outWriter
	startCmd.Stderr = errWriter
	if err := startCmd.Run(); err != nil {
		return result, err
	}

	// Run the command in the sandbox
	cmd := s.execContainerCmd(args)
	cmd.Stdout = outWriter
	cmd.Stderr = errWriter

	if err := cmd.Start(); err != nil {
		return result, err
	}

	err = cmd.Wait()
	if err == nil {
		result.Status = RunStatusSuccess
	} else {
		result.Status = RunStatusFailure
	}

	// Stop the container
	stopCmd := s.stopContainerCmd()
	stopCmd.Stdout = outWriter
	stopCmd.Stderr = errWriter
	if stopErr := stopCmd.Run(); stopErr != nil {
		// Ignore the error if stderr contains "gofer is still running"
		if !strings.Contains(stderr.String(), "gofer is still running") {
			// Don't overwrite the earlier error
			if err == nil {
				err = stopErr
			}
		}
	}

	return result, err
}

// Clean stops and removes all containers.
func (s *podmanSandbox) Clean() error {
	if s.container == "" {
		return nil
	}
	if err := s.forceStopContainer(); err != nil {
		return err
	}
	return podmanCleanContainers()
}
