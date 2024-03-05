package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ossf/package-analysis/internal/log"
)

const (
	podmanBin     = "podman"
	runtimeBin    = "/usr/local/bin/runsc_compat.sh"
	rootDir       = "/var/run/runsc"
	runLogFile    = "runsc.log.boot"
	logDirPattern = "sandbox_logs_"

	// networkName is the name of the podman network defined in
	// tools/network/podman-analysis.conflist. This network is the network
	// used by the sandbox during analysis to separate the sandbox traffic
	// from the host.
	networkName = "analysis-net"
)

type RunStatus uint8

const (
	// RunStatusUnknown is used when some other issue occurred that prevented
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
	status  RunStatus
	stderr  *bytes.Buffer
	stdout  *bytes.Buffer
}

// Log returns the log file recorded during a run.
//
// This log will contain strace data.
func (r *RunResult) Log() (io.ReadCloser, error) {
	return os.Open(r.logPath)
}

func (r *RunResult) Status() RunStatus {
	if r != nil {
		return r.status
	}
	return RunStatusUnknown
}

func (r *RunResult) Stdout() []byte {
	return r.stdout.Bytes()
}

func (r *RunResult) Stderr() []byte {
	return r.stderr.Bytes()
}

type Sandbox interface {
	// Init prepares the sandbox for run and copy commands. The sandbox is
	// only properly initialised if this function returns nil.
	Init(ctx context.Context) error

	// Run executes the supplied command and args in the sandbox.
	// Multiple calls to Run will reuse the same container state,
	// until Clean() is called.
	// The returned RunResult stores information about the execution.
	// If any error occurs, it is returned with a partial RunResult.
	Run(ctx context.Context, command string, args ...string) (*RunResult, error)

	// Clean cleans up the Sandbox. Once called, the Sandbox cannot be used again.
	Clean(ctx context.Context) error

	// CopyIntoSandbox copies a path in the host to one in the sandbox. The paths
	// may be files or directories. The copy fails if the host path does not exist.
	// See https://docs.podman.io/en/latest/markdown/podman-cp.1.html for details
	// on specifying paths.
	// The sandbox must be initialised using Init() before calling this function.
	CopyIntoSandbox(ctx context.Context, hostPath, sandboxPath string) error

	// CopyBackToHost copies a path in the sandbox to one in the host. The paths
	// may be files or directories. The copy fails if the sandbox path does not exist.
	// See https://docs.podman.io/en/latest/markdown/podman-cp.1.html for details
	// on specifying paths.
	// Caution: files coming out of the sandbox are untrusted and proper validation
	// should be performed on the file before use.
	// The sandbox must be initialised using Init() before calling this function.
	CopyBackToHost(ctx context.Context, hostPath, sandboxPath string) error
}

// volume represents a volume mapping between a host src and a container dest.
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
	image       string
	tag         string
	id          string
	container   string
	noPull      bool
	rawSockets  bool
	strace      bool
	offline     bool
	logPackets  bool
	logStdOut   bool
	logStdErr   bool
	echoStdOut  bool
	echoStdErr  bool
	initialised bool
	volumes     []volume
	copies      []copySpec
	environment map[string]string
	logger      *slog.Logger
}

type (
	Option interface{ set(*podmanSandbox) }
	option func(*podmanSandbox) // option implements Option.
)

func (o option) set(sb *podmanSandbox) { o(sb) }

func New(options ...Option) Sandbox {
	sb := &podmanSandbox{
		logger:      slog.Default(),
		environment: make(map[string]string),
	}
	for _, o := range options {
		o.set(sb)
	}

	if sb.image == "" {
		sb.logger.Error("image is required")
		os.Exit(1)
	}
	return sb
}

// Image sets the image to be used by the sandbox. It is a required option.
func Image(image string) Option {
	return option(func(sb *podmanSandbox) { sb.image = image })
}

// EnableRawSockets allows use of raw sockets in the sandbox.
func EnableRawSockets() Option {
	return option(func(sb *podmanSandbox) { sb.rawSockets = true })
}

// EnableStrace enables strace functionality for the sandbox.
func EnableStrace() Option {
	return option(func(sb *podmanSandbox) { sb.strace = true })
}

// Offline disables network functionality for the sandbox.
func Offline() Option {
	return option(func(sb *podmanSandbox) { sb.offline = true })
}

// EnablePacketLogging enables packet logging for the sandbox.
func EnablePacketLogging() Option {
	return option(func(sb *podmanSandbox) { sb.logPackets = true })
}

// LogStdOut enables wrapping each line of stdout from sandboxed process
// as a log.Info line in the main container.
func LogStdOut() Option {
	return option(func(sb *podmanSandbox) { sb.logStdOut = true })
}

// LogStdErr enables wrapping each line of stderr from the sandboxed process
// as log.Warn line in the main container.
func LogStdErr() Option {
	return option(func(sb *podmanSandbox) { sb.logStdErr = true })
}

// EchoStdOut enables simple echoing of the sandboxed process stdout.
func EchoStdOut() Option {
	return option(func(sb *podmanSandbox) { sb.echoStdOut = true })
}

// EchoStdErr enables simple echoing of the sandboxed process stderr.
func EchoStdErr() Option {
	return option(func(sb *podmanSandbox) { sb.echoStdErr = true })
}

// NoPull will disable the image for the sandbox from being pulled during Init.
func NoPull() Option {
	return option(func(sb *podmanSandbox) { sb.noPull = true })
}

// Volume can be used to specify an additional volume map into the container.
// src is the path in the host that will be mapped to the dest path.
func Volume(src, dest string) Option {
	return option(func(sb *podmanSandbox) {
		sb.volumes = append(sb.volumes, volume{
			src:  src,
			dest: dest,
		})
	})
}

// Copy copies a file from the host into the sandbox during initialisation
func Copy(src, dest string) Option {
	return option(func(sb *podmanSandbox) {
		// container ID is set later
		sb.copies = append(sb.copies, hostToContainerCopyCmd(src, dest, ""))
	})
}

func Tag(tag string) Option {
	return option(func(sb *podmanSandbox) { sb.tag = tag })
}

func Logger(logger *slog.Logger) Option {
	return option(func(sb *podmanSandbox) { sb.logger = logger })
}

func SetEnv(key, value string) Option {
	return option(func(sb *podmanSandbox) { sb.environment[key] = value })
}

func removeAllLogs() error {
	matches, err := filepath.Glob(filepath.Join(os.TempDir(), logDirPattern+"*"))
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

func podman(ctx context.Context, args ...string) *exec.Cmd {
	args = append([]string{
		"--cgroup-manager=cgroupfs",
		"--events-backend=file",
	}, args...)
	slog.DebugContext(ctx, "podman", "args", args)
	return exec.CommandContext(ctx, podmanBin, args...)
}

func podmanRun(ctx context.Context, args ...string) error {
	cmd := podman(ctx, args...)
	return cmd.Run()
}

func podmanPrune(ctx context.Context) error {
	return podmanRun(ctx, "image", "prune", "-f")
}

func podmanCleanContainers(ctx context.Context) error {
	return podmanRun(ctx, "rm", "--all", "--force")
}

func (s *podmanSandbox) pullImage(ctx context.Context) error {
	return podmanRun(ctx, "pull", s.imageWithTag())
}

func (s *podmanSandbox) createContainer(ctx context.Context) (string, error) {
	args := []string{
		"create",
		"--runtime=" + runtimeBin,
		"--init",
	}

	networkArgs := []string{
		"--dns=8.8.8.8",  // Manually specify DNS to bypass kube-dns and
		"--dns=8.8.4.4",  // allow for tighter firewall rules that block
		"--dns-search=.", // network traffic to private IP address ranges.
		"--network=" + networkName,
	}

	if s.offline {
		args = append(args, "--network=none")
	} else {
		args = append(args, networkArgs...)
	}

	for k, v := range s.environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, s.extraArgs()...)
	args = append(args, s.imageWithTag())
	cmd := podman(ctx, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(buf.Bytes())), nil
}

func (s *podmanSandbox) startContainerCmd(ctx context.Context, logDir string) *exec.Cmd {
	args := []string{
		"start",
		"--runtime=" + runtimeBin,
		"--runtime-flag=overlay2=none",
		"--runtime-flag=root=" + rootDir,
		"--runtime-flag=debug-log=" + filepath.Join(logDir, "runsc.log.%COMMAND%"),
	}
	if s.rawSockets {
		args = append(args, "--runtime-flag=net-raw")
	}
	if s.strace {
		args = append(args, "--runtime-flag=strace")
	}
	if s.logPackets {
		args = append(args, "--runtime-flag=log-packets")
	}
	args = append(args, s.container)

	return podman(ctx, args...)
}

func (s *podmanSandbox) stopContainerCmd(ctx context.Context) *exec.Cmd {
	return podman(ctx, "stop", s.container)
}

func (s *podmanSandbox) forceStopContainer(ctx context.Context) error {
	return podmanRun(
		ctx,
		"stop",
		"-t=5", // Wait a max of 5 seconds for a graceful stop.
		"-i",   // Ignore any errors of the specified container not being in the store.
		s.container)
}

func (s *podmanSandbox) execContainerCmd(ctx context.Context, execCmd string, execArgs []string) *exec.Cmd {
	args := append([]string{"exec", s.container, execCmd}, execArgs...)
	return podman(ctx, args...)
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

// Init initializes the sandbox, including creating the container and pulling the image.
// The sandbox is marked as initialised if the function completes with no errors.
// If the sandbox has already been marked as initialised, this function simply returns nil.
func (s *podmanSandbox) Init(ctx context.Context) error {
	if s.initialised {
		return nil
	}

	if s.container != "" {
		return nil
	}
	// Delete existing logs (if any).
	if err := removeAllLogs(); err != nil {
		return fmt.Errorf("failed removing all logs: %w", err)
	}
	if err := podmanPrune(ctx); err != nil {
		return fmt.Errorf("error pruning images: %w", err)
	}
	if !s.noPull {
		if err := s.pullImage(ctx); err != nil {
			return fmt.Errorf("error pulling image: %w", err)
		}
	}
	if id, err := s.createContainer(ctx); err != nil {
		return fmt.Errorf("error creating container: %w", err)
	} else {
		s.container = id
	}

	// run each copy command separately
	for _, copyOp := range s.copies {
		copyOp.containerId = s.container
		s.logger.InfoContext(ctx, "podman "+copyOp.String())
		if err := podmanRun(ctx, copyOp.Args()...); err != nil {
			return fmt.Errorf("copy into sandbox [%s]  failed: %w", copyOp, err)
		}
	}

	s.initialised = true

	return nil
}

// Run implements the Sandbox interface.
// If Init() has not yet been run, it will be called automatically before running
func (s *podmanSandbox) Run(ctx context.Context, command string, args ...string) (*RunResult, error) {
	if err := s.Init(ctx); err != nil {
		return &RunResult{}, err
	}

	// Create a place to stash the logs for this run.
	logDir, err := os.MkdirTemp("", logDirPattern)
	if err != nil {
		return &RunResult{}, fmt.Errorf("failed to create log directory: %w", err)
	}
	// Chmod the log dir so it can be read by non-root users. Make the behaviour
	// mimic Mkdir called with 0o777 before umask is applied by applying the
	// umask manually to the permissions.
	umask := syscall.Umask(0)
	syscall.Umask(umask)
	if err := os.Chmod(logDir, fs.FileMode(0o777 & ^umask)); err != nil {
		return &RunResult{}, fmt.Errorf("failed to chmod log directory: %w", err)
	}

	// Prepare the run result.
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	result := &RunResult{
		logPath: filepath.Join(logDir, runLogFile),
		status:  RunStatusUnknown,
		stdout:  &stdout,
		stderr:  &stderr,
	}

	// Prepare stdout and stderr writers
	logOut := log.NewWriter(ctx,
		s.logger.With("command", command, "args", args),
		slog.LevelInfo)
	defer logOut.Close()
	logErr := log.NewWriter(ctx,
		s.logger.With("command", command, "args", args),
		slog.LevelWarn)
	defer logErr.Close()

	outWriters := []io.Writer{&stdout}
	if s.logStdOut {
		outWriters = append(outWriters, logOut)
	}
	if s.echoStdOut {
		outWriters = append(outWriters, os.Stdout)
	}
	outWriter := io.MultiWriter(outWriters...)

	errWriters := []io.Writer{&stderr}
	if s.logStdErr {
		errWriters = append(errWriters, logErr)
	}
	if s.echoStdErr {
		errWriters = append(errWriters, os.Stdout)
	}
	errWriter := io.MultiWriter(errWriters...)

	// Start the container
	startCmd := s.startContainerCmd(ctx, logDir)
	startCmd.Stdout = logOut
	startCmd.Stderr = logErr
	if err := startCmd.Run(); err != nil {
		return result, fmt.Errorf("error starting container: %w", err)
	}

	// Run the command in the sandbox
	cmd := s.execContainerCmd(ctx, command, args)
	cmd.Stdout = outWriter
	cmd.Stderr = errWriter

	if err := cmd.Start(); err != nil {
		return result, fmt.Errorf("error execing command: %w", err)
	}

	err = cmd.Wait()
	if err == nil {
		result.status = RunStatusSuccess
	} else if _, ok := err.(*exec.ExitError); ok {
		result.status = RunStatusFailure
		err = nil
	}

	// Stop the container
	stopCmd := s.stopContainerCmd(ctx)
	var stopStderr bytes.Buffer
	stopCmd.Stdout = logOut
	stopCmd.Stderr = io.MultiWriter(&stopStderr, logErr)
	if stopErr := stopCmd.Run(); stopErr != nil {
		if strings.Contains(stopStderr.String(), "gofer is still running") {
			// Ignore the error if stderr contains "gofer is still running"
			s.logger.DebugContext(ctx, "ignoring 'stop' error - gofer still running")
		} else if err == nil {
			// Don't overwrite the earlier error
			err = fmt.Errorf("error stopping container: %w", stopErr)
		}
	}

	return result, err
}

// Clean implements the Sandbox interface.
func (s *podmanSandbox) Clean(ctx context.Context) error {
	if s.container == "" {
		return nil
	}
	if err := s.forceStopContainer(ctx); err != nil {
		return err
	}
	return podmanCleanContainers(ctx)
}

// CopyIntoSandbox copies a path from the host into the sandbox.
// If the source path does not exist, the command will fail with exit status 125.
func (s *podmanSandbox) CopyIntoSandbox(ctx context.Context, hostPath, sandboxPath string) error {
	if !s.initialised {
		return errors.New("sandbox not initialised")
	}
	if s.container == "" {
		return errors.New("container ID is empty")
	}

	copyCmd := hostToContainerCopyCmd(hostPath, sandboxPath, s.container)
	s.logger.InfoContext(ctx, "podman "+copyCmd.String())
	return podmanRun(ctx, copyCmd.Args()...)
}

// CopyBackToHost copies a path from the sandbox back to the host (after it has run).
// If the source path does not exist, the command will fail with exit status 125.
func (s *podmanSandbox) CopyBackToHost(ctx context.Context, hostPath, sandboxPath string) error {
	if !s.initialised {
		return errors.New("sandbox not initialised")
	}
	if s.container == "" {
		return errors.New("container ID is empty")
	}

	copyCmd := containerToHostCopyCmd(hostPath, sandboxPath, s.container)
	s.logger.InfoContext(ctx, "podman "+copyCmd.String())
	return podmanRun(ctx, copyCmd.Args()...)
}
