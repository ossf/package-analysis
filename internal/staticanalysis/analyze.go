package staticanalysis

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
)

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
			// parsing needed for obfuscation detection
			runTask[Parsing] = true
			runTask[Obfuscation] = true
		default:
			return nil, fmt.Errorf("static analysis task not implemented: %s", task)
		}
	}

	result := Result{}

	if runTask[Basic] {
		result.BasicData = &BasicPackageData{
			Files: map[string]BasicFileData{},
		}
	}

	if runTask[Parsing] {
		result.ParsingData = parsing.PackageResult{}
	}

	walkErr := filepath.WalkDir(extractDir, func(path string, f fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if f.Type().IsRegular() {
			pathInArchive := strings.TrimPrefix(path, extractDir+string(os.PathSeparator))
			log.Info("Processing " + pathInArchive)

			if runTask[Basic] {
				result.BasicData.Files[pathInArchive] = GetBasicFileData(path, pathInArchive)
			}

			if runTask[Parsing] {
				fileParseData, parseErr := parsing.ParseFile(jsParserConfig, path, "", false)
				if parseErr != nil {
					log.Error("Error parsing file", "filename", pathInArchive, "error", parseErr)
					result.ParsingData[pathInArchive] = nil
				} else {
					result.ParsingData[pathInArchive] = fileParseData
				}
			}
		}
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("error while walking package files: %w", walkErr)
	}

	if _, run := runTask[Obfuscation]; run {
		obfuscationData := obfuscation.Analyze(result.ParsingData)
		result.ObfuscationData = &obfuscationData
	}

	return &result, nil

}
