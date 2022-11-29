package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/staticanalysis"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing/js"
	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/internal/worker"
)

var (
	ecosystem   = flag.String("ecosystem", "", "Package ecosystem (required)")
	packageName = flag.String("package", "", "Package name (required)")
	version     = flag.String("version", "", "Package version (ignored if local file is specified)")
	localFile   = flag.String("local", "", "Local package archive containing package to be analysed. Name must match -package argument")
	help        = flag.Bool("help", false, "Prints this help and list of available analyses")
	analyses    = utils.CommaSeparatedFlags("analyses", "obfuscation", "comma-separated list of static analyses to perform")
)

type workDirs struct {
	baseDir    string
	archiveDir string
	extractDir string
	parserDir  string
}

func (wd *workDirs) cleanup() {
	if err := os.RemoveAll(wd.baseDir); err != nil {
		log.Error("Failed to remove work dirs", "baseDir", wd.baseDir, "error", err)
	}
}

const jsParserDirName = "jsparser"

func checkAnalyses(names []string) ([]staticanalysis.Task, error) {
	var tasks []staticanalysis.Task
	for _, name := range names {
		task, ok := staticanalysis.TaskFromString(name)
		if !ok {
			return nil, fmt.Errorf("unrecognised static analysis task: %s", name)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func printAnalyses() {
	println("Available analyses are:")
	for _, task := range staticanalysis.AllTasks() {
		println("\t", task)
	}
}

func doObfuscationDetection(workDirs workDirs) (*obfuscation.AnalysisResult, error) {
	jsParserConfig, err := js.InitParser(path.Join(workDirs.parserDir, jsParserDirName))
	if err != nil {
		return nil, fmt.Errorf("failed to init js parser: %v", err)
	}

	if err != nil {
		return nil, err
	}

	result := &obfuscation.AnalysisResult{
		FileRawData:    map[string]obfuscation.RawData{},
		FileSignals:    map[string]obfuscation.Signals{},
		PackageSignals: obfuscation.NoSignals(),
		ExcludedFiles:  []string{},
	}
	err = filepath.WalkDir(workDirs.extractDir, func(path string, f fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if f.Type().IsRegular() {
			pathInArchive := strings.TrimPrefix(path, workDirs.extractDir+string(os.PathSeparator))
			fmt.Printf("Processing file %s\n", pathInArchive)
			rawData, err := obfuscation.CollectData(jsParserConfig, path, "", false)
			if err != nil {
				log.Error("Error parsing file", "filename", pathInArchive, "error", err)
				result.ExcludedFiles = append(result.ExcludedFiles, pathInArchive)
			} else if rawData == nil {
				// syntax error - could not parse file
				result.ExcludedFiles = append(result.ExcludedFiles, pathInArchive)
			} else {
				// rawData != nil, err == nil
				result.FileRawData[f.Name()] = *rawData
				signals := obfuscation.ComputeSignals(*rawData)
				result.FileSignals[f.Name()] = obfuscation.RemoveNaNs(signals)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error while walking package files: %v", err)
	}

	return result, nil
}

func makeWorkDirs() (workDirs, error) {
	baseDir, err := os.MkdirTemp("", "package-analysis-staticanalyze")
	if err != nil {
		return workDirs{}, err
	}

	archiveDir, err := os.MkdirTemp(baseDir, "archive")
	if err != nil {
		_ = os.RemoveAll(baseDir)
		return workDirs{}, err
	}
	extractDir, err := os.MkdirTemp(baseDir, "extracted")
	if err != nil {
		_ = os.RemoveAll(baseDir)
		return workDirs{}, err
	}
	parserDir, err := os.MkdirTemp(baseDir, "parser")
	if err != nil {
		_ = os.RemoveAll(baseDir)
		return workDirs{}, err
	}

	return workDirs{
		baseDir:    baseDir,
		archiveDir: archiveDir,
		extractDir: extractDir,
		parserDir:  parserDir,
	}, nil
}

func run() (err error) {
	log.Initalize(os.Getenv("LOGGER_ENV"))
	analyses.InitFlag()
	flag.Parse()

	if len(os.Args) == 1 || *help == true {
		flag.Usage()
		println()
		printAnalyses()
		return
	}

	if *ecosystem == "" || *packageName == "" {
		flag.Usage()
		return fmt.Errorf("ecosystem and package are required arguments")
	}

	manager := pkgecosystem.Manager(pkgecosystem.Ecosystem(*ecosystem))
	if manager == nil {
		return fmt.Errorf("unsupported pkg manager for ecosystem %s", *ecosystem)
	}

	pkg, err := worker.ResolvePkg(manager, *packageName, *version, *localFile)
	if err != nil {
		return fmt.Errorf("package error: %v", err)
	}

	uniqueAnalyses := utils.RemoveDuplicates(analyses.Values)
	analysisTasks, err := checkAnalyses(uniqueAnalyses)

	if err != nil {
		printAnalyses()
		return err
	}

	log.Info("Static analysis launched",
		log.Label("package", *packageName),
		log.Label("version", *version),
		log.Label("local_path", *localFile),
		log.Label("analyses", strings.Join(uniqueAnalyses, ",")))

	workDirs, err := makeWorkDirs()
	if err != nil {
		return fmt.Errorf("failed to create work directories: %v", err)
	}

	defer workDirs.cleanup()

	var archivePath string
	if *localFile != "" {
		archivePath = *localFile
	} else {
		archivePath, err = manager.DownloadArchive(pkg.Name(), pkg.Version(), workDirs.archiveDir)
		if err != nil {
			return fmt.Errorf("error downloading archive: %v\n", err)
		}
	}

	err = manager.ExtractArchive(archivePath, workDirs.extractDir)
	if err != nil {
		return fmt.Errorf("archive extraction failed: %v", err)
	}

	for _, task := range analysisTasks {
		switch task {
		case staticanalysis.ObfuscationDetection:
			analysisResult, err := doObfuscationDetection(workDirs)
			if err != nil {
				log.Error("Error occurred during obfuscation detection", "error", err)
			}
			fmt.Printf("Analysis result\n%v\n", analysisResult)
		default:
			return fmt.Errorf("static analysis task not implemented: %s", task)
		}
	}

	return
}

func main() {
	err := run()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
