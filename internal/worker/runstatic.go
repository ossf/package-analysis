package worker

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/staticanalysis"
	"github.com/ossf/package-analysis/internal/utils"
)

const staticAnalysisImage = "gcr.io/ossf-malware-analysis/static-analysis"
const staticAnalyzeBinary = "/usr/local/bin/staticanalyze"
const resultsJSONFile = "/results.json"

func RunStaticAnalyses(pkg *pkgecosystem.Pkg, sbOpts []sandbox.Option, tasks ...staticanalysis.Task) (json.RawMessage, analysis.Status, error) {
	if len(tasks) == 0 {
		tasks = staticanalysis.AllTasks()
	}

	log.Info("Running static analysis tasks", "tasks", tasks)

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

	return resultsJSON, analysis.StatusForRunResult(runResult), nil
}
