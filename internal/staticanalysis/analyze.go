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
AnalyzePackageFiles walks a tree of extracted package files and runs the specified analysis
tasks to produce the result data.

If staticanalysis.Parsing is not in the list of analysisTasks, jsParserConfig may be empty.

If an error occurs while traversing the extracted package directory tree, or an invalid
task is requested, a nil result is returned along with the corresponding error object.
*/
func AnalyzePackageFiles(extractDir string, jsParserConfig parsing.ParserConfig, analysisTasks []Task) (*Result, error) {
	// whether the analysis needs to be run
	runAnalysis := map[Task]bool{}
	// whether the analysis should be output. If a task is not included in analysisTasks
	// but its data is needed for another task, it will be performed but not output
	outputAnalysis := map[Task]bool{}

	for _, task := range analysisTasks {
		switch task {
		case BasicData:
			runAnalysis[BasicData] = true
			outputAnalysis[BasicData] = true
		case Parsing:
			runAnalysis[Parsing] = true
			outputAnalysis[Parsing] = true
		case Obfuscation:
			// parsing needed for obfuscation detection
			runAnalysis[Parsing] = true

			runAnalysis[Obfuscation] = true
			outputAnalysis[Obfuscation] = true
		default:
			return nil, fmt.Errorf("static analysis task not implemented: %s", task)
		}
	}
	basicData := BasicPackageData{
		Files: map[string]BasicFileData{},
	}
	parsingData := parsing.PackageResult{}

	walkErr := filepath.WalkDir(extractDir, func(path string, f fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if f.Type().IsRegular() {
			pathInArchive := strings.TrimPrefix(path, extractDir+string(os.PathSeparator))
			log.Info("Processing " + pathInArchive)

			if runAnalysis[BasicData] {
				basicData.Files[pathInArchive] = GetBasicFileData(path, pathInArchive)
			}

			if runAnalysis[Parsing] {
				fileParseData, parseErr := parsing.ParseFile(jsParserConfig, path, "", false)
				if parseErr != nil {
					log.Error("Error parsing file", "filename", pathInArchive, "error", parseErr)
					parsingData[pathInArchive] = nil
				} else {
					parsingData[pathInArchive] = fileParseData
				}
			}
		}
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("error while walking package files: %w", walkErr)
	}

	var obfuscationData *obfuscation.Result
	if runAnalysis[Obfuscation] {
		obfuscationData = obfuscation.Analyze(parsingData)
	}

	result := Result{}

	if outputAnalysis[BasicData] {
		result[BasicData] = basicData
	}
	if outputAnalysis[Parsing] {
		result[Parsing] = parsingData
	}
	if outputAnalysis[Obfuscation] {
		result[Obfuscation] = obfuscationData
	}

	return &result, nil

}
