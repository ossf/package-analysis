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

// staticAnalysisImage is the Docker image for the static analysis sandbox
const staticAnalysisImage = "gcr.io/ossf-malware-analysis/static-analysis"

// staticAnalyzeBinary is the absolute path to the compiled staticanalyze.go binary
// inside the static analysis sandbox (see sandboxes/staticanalysis/Dockerfile)
const staticAnalyzeBinary = "/usr/local/bin/staticanalyze"

// resultsJSONFile is the absolute path to the shared mount inside the static analysis sandbox
// where the output results JSON data should be written
const resultsJSONFile = "/results.json"

/*
RunStaticAnalysis performs the given static analysis tasks on package code,
in a sandboxed environment. To run all available static analyses, pass
staticanalysis.All as tasks. Use sbOpts to customise sandbox behaviour.
*/
func RunStaticAnalysis(pkg *pkgecosystem.Pkg, sbOpts []sandbox.Option, tasks ...staticanalysis.Task) (result.StaticAnalysisResults, analysis.Status, error) {
	log.Info("Running static analysis", "tasks", tasks)

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
		log.Label("result_status", string(status)),
		"static_analysis_duration", totalTime,
	)

	return resultsJSON, status, nil
}
