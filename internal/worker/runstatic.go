package worker

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/staticanalysis"
	"github.com/ossf/package-analysis/internal/utils"
)

const staticAnalyzeBinary = "/usr/local/bin/staticanalyze"

func RunStaticAnalyses(sb sandbox.Sandbox, pkg *pkgecosystem.Pkg, tasks ...staticanalysis.Task) (json.RawMessage, error) {
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
	}

	if pkg.IsLocal() {
		args = append(args, "-local", pkg.LocalPath())
	}

	r, err := sb.Run(args...)
	if err != nil {
		return nil, fmt.Errorf("sandbox failed (%w)", err)
	}

	return r.Stdout(), nil
}
