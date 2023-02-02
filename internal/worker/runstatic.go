package worker

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/staticanalysis"
	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/pkg/result"
)

const staticAnalysisImage = "gcr.io/ossf-malware-analysis/static-analysis"
const staticAnalyzeBinary = "/usr/local/bin/staticanalyze"
const resultsJSONFile = "/results.json"

/*
RunStaticAnalyses performs sandboxed static analysis on the given package.
Use sbOpts to customise sandbox behaviour.
If len(tasks) > 0, only the specified static analysis tasks will be performed.
Otherwise, all available tasks [staticanalysis.AllTasks()] will be performed.
*/
func RunStaticAnalyses(pkg *pkgecosystem.Pkg, sbOpts []sandbox.Option, tasks ...staticanalysis.Task) (result.StaticAnalysisResults, analysis.Status, error) {
	if len(tasks) == 0 {
		tasks = staticanalysis.AllTasks()
	}

	log.Info("Running static analysis tasks", "tasks", tasks)

	startTime := time.Now()

	analyses := utils.Transform(tasks, func(t staticanalysis.Task) string { return string(t) })

	args := []string{
		staticAnalyzeBinary,
		"-ecosystem", pkg.EcosystemName(),
		"-package", pkg.Name(),
		"-version", pkg.Version(),
		"-analyses", strings.Join(analyses, ","),
		"-output", resultsJSONFile,
	}

	if pkg.IsLocal() {
		args = append(args, "-local", pkg.LocalPath())
	}

	// create the results JSON file as an empty file, so it can be mounted into the container
	resultsFile, err := os.OpenFile(resultsJSONFile, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("could not create results JSON file: %v", err)
	}
	_ = resultsFile.Close()

	// for saving static analysis results inside the sandbox
	sbOpts = append(sbOpts, sandbox.Volume(resultsJSONFile, resultsJSONFile))

	sb := sandbox.New(staticAnalysisImage, sbOpts...)
	defer func() {
		if err := sb.Clean(); err != nil {
			log.Error("error cleaning up sandbox", "error", err)
		}
	}()

	runResult, err := sb.Run(args...)
	if err != nil {
		return nil, "", fmt.Errorf("sandbox failed (%w)", err)
	}

	resultsJSON, err := os.ReadFile(resultsJSONFile)
	if err != nil {
		return nil, "", fmt.Errorf("could not read results JSON file: %v", err)
	}

	log.Info("Got results", "length", len(resultsJSON))

	status := analysis.StatusForRunResult(runResult)

	totalTime := time.Since(startTime)
	log.Info("Static analysis finished",
		log.Label("ecosystem", pkg.EcosystemName()),
		"name", pkg.Name(),
		"version", pkg.Version(),
		"result_status", string(status),
		"static_analysis_duration_sec", fmt.Sprintf("%.1f", totalTime.Seconds()),
	)

	return resultsJSON, status, nil
}
