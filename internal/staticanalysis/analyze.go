package staticanalysis

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/staticanalysis/basicdata"
	"github.com/ossf/package-analysis/internal/staticanalysis/externalcmd"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/signals"
)

// enumeratePackageFiles returns a list of absolute paths to all (regular) files
// in a directory or any descendent directory.
func enumeratePackageFiles(extractDir string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(extractDir, func(path string, f fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if f.Type().IsRegular() {
			paths = append(paths, path)
		}
		return nil
	})

	return paths, err
}

/*
AnalyzePackageFiles walks a tree of extracted package files and runs the analysis tasks
listed in analysisTasks to produce the result data.

Note that to some tasks depend on the data from other tasks; for example, 'signals'
depends on 'parsing'. If a task listed in analysisTasks depends on a task not listed
in analysisTasks, then both tasks are performed.

If staticanalysis.Parsing is not in the list of analysisTasks, jsParserConfig may be empty.

If an error occurs while traversing the extracted package directory tree, or an invalid
task is requested, a nil result is returned along with the corresponding error object.
*/
func AnalyzePackageFiles(extractDir string, jsParserConfig parsing.ParserConfig, analysisTasks []Task) (*Result, error) {
	runTask := map[Task]bool{}

	for _, task := range analysisTasks {
		switch task {
		case Basic:
			runTask[Basic] = true
		case Parsing:
			runTask[Parsing] = true
		case Signals:
			if !runTask[Parsing] {
				log.Info("adding staticanalysis.Parsing to task list (needed by staticanalysis.Signals)")
			}
			runTask[Parsing] = true
			runTask[Signals] = true
		case All:
			return nil, errors.New("staticanalysis.All should not be passed in directly, use staticanalysis.AllTasks() instead")
		default:
			return nil, fmt.Errorf("static analysis task not implemented: %s", task)
		}
	}

	paths, err := enumeratePackageFiles(extractDir)
	if err != nil {
		return nil, fmt.Errorf("error enumerating package files: %w", err)
	}

	getPathInArchive := func(absolutePath string) string {
		return strings.TrimPrefix(absolutePath, extractDir+string(os.PathSeparator))
	}
	// inverse of above function
	getAbsolutePath := func(packagePath string) string {
		return filepath.Join(extractDir, packagePath)
	}

	fileResults := make([]SingleResult, 0, len(paths))
	for _, path := range paths {
		fileResults = append(fileResults, SingleResult{Filename: getPathInArchive(path)})
	}

	if runTask[Basic] {
		log.Info("run basic analysis")
		basicData, err := basicdata.Analyze(paths, getPathInArchive)
		if err != nil {
			log.Error("static analysis basic data error", "error", err)
		} else if len(basicData) != len(fileResults) {
			log.Error(fmt.Sprintf("basicdata.Analyze() returned %d results, expecting %d",
				len(basicData), len(fileResults)), log.Label("task", string(Basic)))
		} else {
			for i := range fileResults {
				fileResults[i].Basic = &basicData[i]
			}
		}
	}

	if runTask[Parsing] {
		log.Info("run parsing analysis")

		input := externalcmd.MultipleFileInput(paths)
		parsingResults, err := parsing.Analyze(jsParserConfig, input, false)

		if err != nil {
			log.Error("static analysis parsing error", "error", err)
		} else if len(parsingResults) != len(fileResults) {
			log.Error(fmt.Sprintf("parsing.Analyze() returned %d results, expecting %d",
				len(parsingResults), len(fileResults)), log.Label("task", string(Basic)))
		} else {
			for i, r := range fileResults {
				fileParseResult := parsingResults[getAbsolutePath(r.Filename)]
				fileResults[i].Parsing = &fileParseResult
			}
		}
	}

	if runTask[Signals] {
		log.Info("run signals analysis")
		for i, r := range fileResults {
			if r.Parsing != nil {
				singleData := signals.AnalyzeSingle(*r.Parsing)
				fileResults[i].Signals = &singleData
			} else {
				log.Warn("skipped signals analysis due to no parsing data", "filename", r.Filename)
			}
		}
	}

	return &Result{Files: fileResults}, nil
}
