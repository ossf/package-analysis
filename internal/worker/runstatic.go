package worker

import (
	"fmt"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/staticanalysis"
)

type StaticAnalysisResults map[staticanalysis.Task]any

const staticAnalyzeBinary = "/usr/local/bin/staticanalyze"

func RunStaticAnalyses(sb sandbox.Sandbox, pkg *pkgecosystem.Pkg, tasks ...staticanalysis.Task) (results StaticAnalysisResults, err error) {
	if len(tasks) == 0 {
		tasks = staticanalysis.AllTasks()
	}

	log.Info("Running static analysis tasks",
		"tasks", tasks)

	args := []string{
		staticAnalyzeBinary,
		"-ecosystem", pkg.EcosystemName(),
		"-package", pkg.Name(),
		"-version", pkg.Version(),
	}

	if pkg.IsLocal() {
		return nil, fmt.Errorf("local packages are not yet supported")
	}

	r, err := sb.Run(args...)

	if err != nil {
		return nil, fmt.Errorf("sandbox failed (%w)", err)
	}
	results = make(StaticAnalysisResults)

	results[staticanalysis.ObfuscationDetection] = r.Stdout()

	return results, err
}
