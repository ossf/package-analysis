package worker

import (
	"fmt"
	"os"
	"time"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/pkg/result"
)

const staticAnalysisImage = "gcr.io/ossf-malware-analysis/static-analysis"
const staticAnalyzeBinary = "/usr/local/bin/staticanalyze"
const resultsJSONFile = "/results.json"

// RunStaticAnalyses performs sandboxed static analysis on the given package.
// Use sbOpts to customise sandbox behaviour.
func RunStaticAnalyses(pkg *pkgecosystem.Pkg, sbOpts []sandbox.Option) (result.StaticAnalysisResults, analysis.Status, error) {
	log.Info("Running static analysis")

	startTime := time.Now()

	args := []string{
		staticAnalyzeBinary,
		"-ecosystem", pkg.EcosystemName(),
		"-package", pkg.Name(),
		"-version", pkg.Version(),
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
