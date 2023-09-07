package staticanalysis

import (
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
			// ignore
		default:
			return nil, fmt.Errorf("static analysis task not implemented: %s", task)
		}
	}

	fileList, err := enumeratePackageFiles(extractDir)
	if err != nil {
		return nil, fmt.Errorf("error enumerating package files: %w", err)
	}

	result := Result{}

	getPathInArchive := func(absolutePath string) string {
		return strings.TrimPrefix(absolutePath, extractDir+string(os.PathSeparator))
	}

	if runTask[Basic] {
		log.Info("run basic analysis")
		basicData, err := basicdata.Analyze(fileList, getPathInArchive)
		if err != nil {
			log.Error("static analysis error", log.Label("task", string(Basic)), "error", err)
		} else {
			result.BasicData = basicData
		}
	}

	if runTask[Parsing] {
		log.Info("run parsing analysis")

		input := externalcmd.MultipleFileInput(fileList)
		parsingResults, err := parsing.Analyze(jsParserConfig, input, false)

		if err != nil {
			log.Error("static analysis error", log.Label("task", string(Parsing)), "error", err)
		} else {
			// change absolute path in parsingResults to package-relative path
			for i, r := range parsingResults {
				parsingResults[i].Filename = getPathInArchive(r.Filename)
			}
			result.ParsingData = parsingResults
		}
	}

	if runTask[Signals] {
		if len(result.ParsingData) > 0 {
			log.Info("run signals analysis")

			signalsData := signals.Analyze(result.ParsingData)
			result.SignalsData = &signalsData
		} else {
			log.Warn("skipped signals analysis due to no parsing data")
		}
	}

	return &result, nil
}
