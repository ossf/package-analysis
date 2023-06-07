package worker

import (
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/dynamicanalysis"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

func retrieveExecutionLog(sb sandbox.Sandbox) (string, error) {
	// retrieve execution log back to host
	executionLogDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	defer os.RemoveAll(executionLogDir)
	executionLogPath := path.Join(executionLogDir, "execution.log")

	// if the copy fails, it could be that the execution log is not actually present.
	// For now, we'll just log the error and otherwise ignore it
	if err := sb.CopyToHost("/execution.log", executionLogPath); err != nil {
		log.Warn("Could not copy execution log from sandbox", "error", err)
		return "", nil
	}

	logData, err := os.ReadFile(executionLogPath)
	if err != nil {
		return "", err
	}

	// remove control characters except tab (\x09) and newline (\x0A)
	processedLog := nonSpaceControlChars.ReplaceAllLiteral(logData, []byte{})
	log.Info("Read execution log", "rawLength", len(logData), "processedLength", len(processedLog))

	return string(processedLog), nil
}

/*
DynamicAnalysisResult holds all the results from RunDynamicAnalysis

AnalysisData: Map of each successfully run phase to a summary of
the corresponding dynamic analysis result. This summary has two parts:
1. StraceSummary: information about system calls performed by the process
2. FileWrites: list of files which were written to and counts of bytes written

Note, if error is not nil, then results[lastRunPhase] is nil.

LastRunPhase: the last phase that was run. If error is non-nil, this phase did not
successfully complete, and the results for this phase are not recorded.
Otherwise, the results contain data for this phase, even in cases where the
sandboxed process terminated abnormally.

Status: the status of the last run phase if it completed without error, else empty
*/

type DynamicAnalysisResult struct {
	Data         analysisrun.DynamicAnalysisResults
	LastRunPhase analysisrun.DynamicPhase
	LastStatus   analysis.Status
}

// defaultDynamicAnalysisImage is container image name of the default dynamic analysis sandbox
const defaultDynamicAnalysisImage = "gcr.io/ossf-malware-analysis/dynamic-analysis"

// defaultDynamicAnalysisCommand returns the path (in the default sandbox image)
// of the default dynamic analysis command for the ecosystem
var defaultDynamicAnalysisCommand = map[pkgecosystem.Ecosystem]string{
	pkgecosystem.CratesIO:  "/usr/local/bin/analyze-rust.py",
	pkgecosystem.NPM:       "/usr/local/bin/analyze-node.js",
	pkgecosystem.Packagist: "/usr/local/bin/analyze-php.php",
	pkgecosystem.PyPI:      "/usr/local/bin/analyze-python.py",
	pkgecosystem.RubyGems:  "/usr/local/bin/analyze-ruby.rb",
}

// execArgs returns the arguments to pass to the dynamic analysis command in the sandbox
// for the given phase of dynamic analysis on a package. The actual analysis command
// depends on the ecosystem, see pkgmanager.PkgManager.DynamicAnalysisCommand()
func analysisArgs(p *pkgmanager.Pkg, phase analysisrun.DynamicPhase) []string {
	args := make([]string, 0)

	if p.IsLocal() {
		args = append(args, "--local", p.LocalPath())
	} else if p.Version() != "" {
		args = append(args, "--version", p.Version())
	}

	if phase == "" {
		args = append(args, "all")
	} else {
		args = append(args, string(phase))
	}

	args = append(args, p.Name())

	return args
}

var nonSpaceControlChars = regexp.MustCompile("[\x00-\x08\x0b-\x1f\x7f]")

/*
RunDynamicAnalysis runs dynamic analysis on the given package in the sandbox
provided, across various phases (e.g. import, install) valid in the package ecosystem.
Status and errors are logged to stdout.

All data and status relating to analysis (including errors produced by invalid packages)
is returned in the DynamicAnalysisResult struct.

The error value holds any error that occurred in the runtime/sandbox infrastructure,
excluding from within the analysis itself. In other words, it does not include errors
produced by the package under analysis.
*/

func RunDynamicAnalysis(pkg *pkgmanager.Pkg, sbOpts []sandbox.Option) (DynamicAnalysisResult, error) {
	var beforeDynamic runtime.MemStats
	runtime.ReadMemStats(&beforeDynamic)
	log.Info("Memory Stats, heap usage before dynamic analysis",
		log.Label("ecosystem", pkg.EcosystemName()),
		"name", pkg.Name(),
		"version", pkg.Version(),
		"heap_usage_before_dynamic_analysis", strconv.FormatUint(beforeDynamic.Alloc, 10),
	)

	data := analysisrun.DynamicAnalysisResults{
		StraceSummary:      make(analysisrun.DynamicAnalysisStraceSummary),
		FileWritesSummary:  make(analysisrun.DynamicAnalysisFileWritesSummary),
		FileWriteBufferIds: make(analysisrun.DynamicAnalysisFileWriteBufferIds),
	}

	sb := sandbox.New(sbOpts...)

	defer func() {
		if err := sb.Clean(); err != nil {
			log.Error("error cleaning up sandbox", "error", err)
		}
	}()

	var lastRunPhase analysisrun.DynamicPhase
	var lastStatus analysis.Status
	var lastError error
	for _, phase := range analysisrun.DefaultDynamicPhases() {
		startTime := time.Now()
		phaseResult, err := dynamicanalysis.Run(sb, analysisArgs(pkg, phase))
		lastRunPhase = phase

		runDuration := time.Since(startTime)
		log.Info("Dynamic analysis phase finished",
			log.Label("ecosystem", pkg.EcosystemName()),
			"name", pkg.Name(),
			"version", pkg.Version(),
			log.Label("phase", string(phase)),
			"error", err,
			"dynamic_analysis_phase_duration", runDuration,
		)

		if err != nil {
			// Error when trying to actually run; don't record the result for this phase
			// or attempt subsequent phases
			lastStatus = ""
			lastError = err
			break
		}

		data.StraceSummary[phase] = &phaseResult.StraceSummary
		data.FileWritesSummary[phase] = &phaseResult.FileWritesSummary
		lastStatus = phaseResult.StraceSummary.Status
		data.FileWriteBufferIds[phase] = phaseResult.FileWriteBufferIds

		if lastStatus != analysis.StatusCompleted {
			// Error caused by an issue with the package (probably).
			// Don't continue with phases if this one did not complete successfully.
			break
		}
	}

	var afterDynamic runtime.MemStats
	runtime.ReadMemStats(&afterDynamic)
	log.Info("Memory Stats, heap usage after dynamic analysis",
		log.Label("ecosystem", pkg.EcosystemName()),
		"name", pkg.Name(),
		"version", pkg.Version(),
		"heap_usage_after_dynamic_analysis", strconv.FormatUint(afterDynamic.Alloc, 10))

	if lastError != nil {
		LogDynamicAnalysisError(pkg, lastRunPhase, lastError)
		return DynamicAnalysisResult{data, lastRunPhase, lastStatus}, lastError
	}

	LogDynamicAnalysisResult(pkg, lastRunPhase, lastStatus)

	executionLog, err := retrieveExecutionLog(sb)
	if err != nil {
		// don't return this error, just log it
		log.Error("Error retrieving execution log", "error", err)
	} else {
		data.ExecutionLog = analysisrun.DynamicAnalysisExecutionLog(executionLog)
	}

	return DynamicAnalysisResult{data, lastRunPhase, lastStatus}, nil
}
