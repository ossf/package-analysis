package sandbox

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/ossf/package-analysis/internal/log"
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

type Sandbox interface {
	// Init prepares the Sandbox ready for running commands.
	//
	// It must be called before Run so that any setup work is complete before
	// any commands are executed.
	Init() error

	// Run will run the sandbox for the given args.
	//
	// The container used to execute the command will be removed when the command
	// is completed.
	Run(...string) (*RunResult, error)
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
	image   string
	tag     string
	init    bool
	noPull  bool
	volumes []volume
}

type Option interface{ set(*podmanSandbox) }
type option func(*podmanSandbox)       // option implements Option.
func (o option) set(sb *podmanSandbox) { o(sb) }

func New(image string, options ...Option) Sandbox {
	sb := &podmanSandbox{
		image:   image,
		tag:     "",
		init:    false,
		noPull:  false,
		volumes: make([]volume, 0),
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

func podmanRunCmd(image string, commandArgs []string, extraArgs []string) *exec.Cmd {
	args := []string{
		"run",
		"--runtime=" + runtimeBin,
		"--runtime-flag=root=" + rootDir,
		"--runtime-flag=debug-log=/tmp/runsc.log.%COMMAND%",
		"--runtime-flag=net-raw",
		"--runtime-flag=strace",
		"--runtime-flag=log-packets",
		"--cgroup-manager=cgroupfs",
		"--events-backend=file",
		"--hostname=box",
		"--rm",
	}
	args = append(args, extraArgs...)
	args = append(args, image)
	args = append(args, commandArgs...)

	cmd := exec.Command(podmanBin, args...)
	return cmd
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

// Initializes the Sandbox ready for running commands.
//
// The image supplied will be pulled if it hasn't already been.
func (s *podmanSandbox) Init() error {
	if s.init {
		return nil
	}
	if !s.noPull {
		if err := podmanPull(s.imageWithTag()); err != nil {
			return err
		}
	}
	if err := podmanPrune(); err != nil {
		return err
	}
	s.init = true
	return nil
}

// Run will run a single command inside the sandbox.
//
// The container used to execute the command will be removed when the command
// is completed.
//
// This function is useful for running multiple commands in the sandbox.
func (s *podmanSandbox) Run(args ...string) (*RunResult, error) {
	// Delete existing logs (if any).
	// This function uses a fixed log name and is not threadsafe.
	if err := os.RemoveAll(logPath); err != nil {
		return &RunResult{}, err
	}

	cmd := podmanRunCmd(s.imageWithTag(), args, s.extraArgs())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	logOut := log.Writer(log.InfoLevel,
		"args", args)
	defer logOut.Close()
	logErr := log.Writer(log.WarnLevel,
		"args", args)
	defer logErr.Close()
	cmd.Stdout = io.MultiWriter(&stdout, logOut)
	cmd.Stderr = io.MultiWriter(&stderr, logErr)

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
func (s *podmanSandbox) Clean() error {
	// remove log files too?
	return podmanCleanContainers()
}
