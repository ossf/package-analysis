package staticanalysis

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/staticanalysis/externalcmd"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
)

// enumeratePackageFiles returns a list of absolute paths to all (regular) files
// in a directory or any descendent directory
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

func getPathInArchive(path, extractDir string) string {
	return strings.TrimPrefix(path, extractDir+string(os.PathSeparator))
}

/*
AnalyzePackageFiles walks a tree of extracted package files and runs the analysis tasks
listed in analysisTasks to produce the result data.

Note that to some tasks depend on the data from other tasks; for example, 'obfuscation'
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
		case Obfuscation:
			if !runTask[Parsing] {
				log.Info("adding staticanalysis.Parsing to task list (needed by staticanalysis.Obfuscation)")
			}
			runTask[Parsing] = true
			runTask[Obfuscation] = true
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

	archivePath := map[string]string{}
	for _, path := range fileList {
		archivePath[path] = getPathInArchive(path, extractDir)
	}

	if runTask[Basic] {
		log.Info("run basic analysis")
		basicData, err := GetBasicData(fileList, archivePath)
		if err != nil {
			return nil, fmt.Errorf("error when collecting basic data: %w", err)
		}
		result.BasicData = basicData
	}

	if runTask[Parsing] {
		log.Info("run parsing analysis")

		input := externalcmd.MultipleFileInput(fileList)
		parsingResults, err := parsing.Analyze(jsParserConfig, input, false)

		result.ParsingData = parsing.PackageResult{}
		if err != nil {
			return nil, fmt.Errorf("error while parsing package files: %w", err)
		}
		for path, parseResult := range parsingResults {
			pathInArchive := archivePath[path]
			result.ParsingData[pathInArchive] = parseResult
		}
	}

	if runTask[Obfuscation] {
		log.Info("run obfuscation analysis")

		obfuscationData := obfuscation.Analyze(result.ParsingData)
		result.ObfuscationData = &obfuscationData
	}

	return &result, nil
}
