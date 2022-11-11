package worker

import (
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/staticanalysis"
)

type StaticAnalysisResults map[staticanalysis.Task]any

func RunStaticAnalyses(sb sandbox.Sandbox, pkg *pkgecosystem.Pkg, tasks ...staticanalysis.Task) (results StaticAnalysisResults, err error) {
	if len(tasks) == 0 {
		tasks = staticanalysis.AllTasks()
	}

	results = make(StaticAnalysisResults)

	// TODO implement

	return results, err
}
