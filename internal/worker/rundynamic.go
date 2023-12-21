package worker

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	mathrand "math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/dynamicanalysis"
	"github.com/ossf/package-analysis/internal/featureflags"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

// defaultDynamicAnalysisImage is container image name of the default dynamic analysis sandbox
const defaultDynamicAnalysisImage = "gcr.io/ossf-malware-analysis/dynamic-analysis"

/*
DynamicAnalysisResult holds all data and status from RunDynamicAnalysis.

Data: analysisrun.DynamicAnalysisData for the package under analysis.
Note, if error is not nil, then results[lastRunPhase] is nil.

LastRunPhase: the last phase that was run. If error is non-nil, this phase did not
successfully complete, and the results for this phase are not recorded.
Otherwise, the results contain data for this phase, even in cases where the
sandboxed process terminated abnormally.

LastStatus: the status of the last run phase if it completed without error, else empty
*/

type DynamicAnalysisResult struct {
	Data         analysisrun.DynamicAnalysisData
	LastRunPhase analysisrun.DynamicPhase
	LastStatus   analysis.Status
}

func dynamicPhases(ecosystem pkgecosystem.Ecosystem) []analysisrun.DynamicPhase {
	phases := analysisrun.DefaultDynamicPhases()

	// currently, the execute phase is only supported for python analysis
	executePhaseSupported := map[pkgecosystem.Ecosystem]struct{}{
		pkgecosystem.PyPI: {},
	}

	if featureflags.CodeExecution.Enabled() {
		if _, supported := executePhaseSupported[ecosystem]; supported {
			phases = append(phases, analysisrun.DynamicPhaseExecute)
		}
	}

	return phases
}

// addSSHKeysToSandbox generates a new rsa private and public key pair
// and copies them into the ~/.ssh directory of the sandbox with the
// default file names.
func addSSHKeysToSandbox(ctx context.Context, sb sandbox.Sandbox) error {
	generatedPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	generatedPublicKey := generatedPrivateKey.PublicKey
	if err != nil {
		return err
	}

	tempdir, err := os.MkdirTemp(".", "temp_ssh_dir")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempdir)
	privateKeyFile, err := os.Create(filepath.Join(tempdir, "id_rsa"))
	if err != nil {
		return err
	}
	defer privateKeyFile.Close()
	pubKeyFile, err := os.Create(filepath.Join(tempdir, "id_rsa.pub"))
	if err != nil {
		return err
	}
	defer pubKeyFile.Close()

	openSSHPrivateKeyBlock, err := ssh.MarshalPrivateKey(generatedPrivateKey, "")
	if err = pem.Encode(privateKeyFile, openSSHPrivateKeyBlock); err != nil {
		return err
	}
	publicKey, err := ssh.NewPublicKey(&generatedPublicKey)
	if err != nil {
		return err
	}
	pubKeyFile.Write(ssh.MarshalAuthorizedKey(publicKey))
	return sb.CopyIntoSandbox(ctx, tempdir+"/.", "/root/.ssh")
}

// generateAWSKeys returns two strings. The first is an AWS access key id based
// off of some known patterns and pseudorandom values. The second is a random 30
// byte base64 encoded string to use as an AWS secret access key.
func generateAWSKeys() (string, string) {
	const charSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	var accessKeyId = "AKIAI"
	src := mathrand.NewSource(time.Now().UnixNano())
	r := mathrand.New(src)
	for i := 0; i < 14; i++ {
		randIndex := r.Intn(len(charSet))
		accessKeyId += string(charSet[randIndex])
	}
	accessKeyId += "Q"
	b := make([]byte, 30)
	r.Read(b)
	return accessKeyId, base64.StdEncoding.EncodeToString(b)
}

/*
RunDynamicAnalysis runs dynamic analysis on the given package across the phases
valid in the package ecosystem (e.g. import, install), in a sandbox created
using the provided options. The options must specify the sandbox image to use.

analysisCmd is an optional argument used to override the default command run
inside the sandbox to perform the analysis. It must support the interface
described under "Adding a new Runtime Analysis script" in sandboxes/README.md

All data and status relating to analysis (including errors produced by invalid packages)
is returned in the DynamicAnalysisResult struct. Status and errors are also logged to stdout.

The returned error holds any error that occurred in the runtime/sandbox infrastructure,
excluding from within the analysis itself. In other words, it does not include errors
produced by the package under analysis.
*/
func RunDynamicAnalysis(ctx context.Context, pkg *pkgmanager.Pkg, sbOpts []sandbox.Option, analysisCmd string) (DynamicAnalysisResult, error) {
	ctx = log.ContextWithAttrs(ctx, slog.String("mode", "dynamic"))

	var beforeDynamic runtime.MemStats
	runtime.ReadMemStats(&beforeDynamic)
	slog.InfoContext(ctx, "Memory Stats, heap usage before dynamic analysis",
		"heap_usage_before_dynamic_analysis", strconv.FormatUint(beforeDynamic.Alloc, 10),
	)

	if analysisCmd == "" {
		analysisCmd = dynamicanalysis.DefaultCommand(pkg.Ecosystem())
	}

	// Adding environment variable baits. We use mocked AWS keys since they are
	// commonly added as environment variables and will be easy to query for in
	// the analysis results. See AWS docs on environment variable configuration:
	// https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html
	AWSAccessKeyId, AWSSecretAccessKey := generateAWSKeys()
	sbOpts = append(sbOpts, sandbox.SetEnv("AWS_ACCESS_KEY_ID", AWSAccessKeyId))
	sbOpts = append(sbOpts, sandbox.SetEnv("AWS_SECRET_ACCESS_KEY", AWSSecretAccessKey))

	sb := sandbox.New(sbOpts...)

	defer func() {
		if err := sb.Clean(ctx); err != nil {
			slog.ErrorContext(ctx, "Error cleaning up sandbox", "error", err)
		}
	}()

	// initialise sandbox before copy/run
	if err := sb.Init(ctx); err != nil {
		LogDynamicAnalysisError(ctx, pkg, "", err)
		return DynamicAnalysisResult{}, err
	}

	if err := addSSHKeysToSandbox(ctx, sb); err != nil {
		// Log error and proceed without ssh keys.
		LogDynamicAnalysisError(ctx, pkg, "", err)
	}

	result := DynamicAnalysisResult{
		Data: analysisrun.DynamicAnalysisData{
			StraceSummary:      make(analysisrun.DynamicAnalysisStraceSummary),
			FileWritesSummary:  make(analysisrun.DynamicAnalysisFileWritesSummary),
			FileWriteBufferIds: make(analysisrun.DynamicAnalysisFileWriteBufferIds),
		},
	}

	// lastError holds the error that occurred in the most recently run dynamic analysis phase.
	// This is not a part of the result because a non-nil value means that the error originated
	// from our code, as opposed to the package under analysis
	var lastError error

	for _, phase := range dynamicPhases(pkg.Ecosystem()) {
		if err := runDynamicAnalysisPhase(ctx, pkg, sb, analysisCmd, phase, &result); err != nil {
			// Error when trying to actually run; don't record the result for this phase
			// or attempt subsequent phases
			result.LastStatus = ""
			lastError = err
			break
		}

		if result.LastStatus != analysis.StatusCompleted {
			// Error caused by an issue with the package (probably).
			// Don't continue with phases if this one did not complete successfully.
			break
		}
	}

	var afterDynamic runtime.MemStats
	runtime.ReadMemStats(&afterDynamic)
	slog.InfoContext(ctx, "Memory Stats, heap usage after dynamic analysis",
		"heap_usage_after_dynamic_analysis", strconv.FormatUint(afterDynamic.Alloc, 10))

	if lastError != nil {
		LogDynamicAnalysisError(ctx, pkg, result.LastRunPhase, lastError)
		return result, lastError
	}

	LogDynamicAnalysisResult(ctx, pkg, result.LastRunPhase, result.LastStatus)

	return result, nil
}

// openStraceDebugLogFile creates and returns the file to be used for debug logging of strace parsing
// during a dynamic analysis phase. The file is created with the given filename in log.StraceDebugLogDir.
// It is truncated on open (so a unique name per analysis phase should be used) and is the caller's
// responsibility to close. If strace debug logging is disabled, or some error occurs during creation,
// a nil file pointer is returned, and nothing more need be done by the caller.
func openStraceDebugLogFile(ctx context.Context, name string) *os.File {
	if !featureflags.StraceDebugLogging.Enabled() {
		return nil
	}

	var logDir = log.StraceDebugLogDir
	if err := os.MkdirAll(logDir, 0o777); err != nil {
		slog.WarnContext(ctx, "could not create directory for strace debug logs", "path", logDir, "error", err)
	}

	logPath := filepath.Join(logDir, name)
	if logFile, err := os.Create(logPath); err != nil {
		slog.WarnContext(ctx, "could not create strace debug log file", "path", logPath, "error", err)
		return nil
	} else {
		return logFile
	}
}

func straceDebugLogFilename(pkg *pkgmanager.Pkg, phase analysisrun.DynamicPhase) string {
	filename := fmt.Sprintf("%s-%s", pkg.Ecosystem(), pkg.Name())
	if pkg.Version() != "" {
		filename += "-" + pkg.Version()
	}
	filename += fmt.Sprintf("-%s-strace.log", phase)

	// Protect against e.g. a package name that contains a slash.
	// This may cause name collisions, but it's probably fine for a debug log
	return strings.ReplaceAll(filename, string(os.PathSeparator), "-")
}

func runDynamicAnalysisPhase(ctx context.Context, pkg *pkgmanager.Pkg, sb sandbox.Sandbox, analysisCmd string, phase analysisrun.DynamicPhase, result *DynamicAnalysisResult) error {
	phaseCtx := log.ContextWithAttrs(ctx, log.Label("phase", string(phase)))
	startTime := time.Now()
	args := dynamicanalysis.MakeAnalysisArgs(pkg, phase)

	straceLogger := slog.New(slog.NewTextHandler(io.Discard, nil)) // default is nop logger
	if logFile := openStraceDebugLogFile(phaseCtx, straceDebugLogFilename(pkg, phase)); logFile != nil {
		slog.InfoContext(phaseCtx, "strace debug logging enabled")
		defer logFile.Close()

		enableDebug := &slog.HandlerOptions{Level: slog.LevelDebug}
		straceLogger = slog.New(log.NewContextLogHandler(slog.NewTextHandler(logFile, enableDebug)))
		straceLogger.InfoContext(phaseCtx, "running dynamic analysis")
	}

	phaseResult, err := dynamicanalysis.Run(phaseCtx, sb, analysisCmd, args, straceLogger)
	result.LastRunPhase = phase
	runDuration := time.Since(startTime)
	slog.InfoContext(phaseCtx, "Dynamic analysis phase finished",
		"error", err,
		"dynamic_analysis_phase_duration", runDuration,
	)

	if err != nil {
		return err
	}

	result.Data.StraceSummary[phase] = &phaseResult.StraceSummary
	result.Data.FileWritesSummary[phase] = &phaseResult.FileWritesSummary
	result.Data.FileWriteBufferIds[phase] = phaseResult.FileWriteBufferIds
	result.LastStatus = phaseResult.StraceSummary.Status

	if phase == analysisrun.DynamicPhaseExecute {
		executionLog, err := retrieveExecutionLog(ctx, sb)
		if err != nil {
			// don't return this error, just log it
			slog.ErrorContext(ctx, "Error retrieving execution log", "error", err)
		} else {
			result.Data.ExecutionLog = analysisrun.DynamicAnalysisExecutionLog(executionLog)
		}
	}

	return nil
}
