package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/staticanalysis"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/internal/worker"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

var (
	ecosystem   pkgecosystem.Ecosystem
	packageName = flag.String("package", "", "package name (required)")
	version     = flag.String("version", "", "package version (ignored if local file is specified)")
	localFile   = flag.String("local", "", "local package archive containing package to be analysed. Name must match -package argument")
	output      = flag.String("output", "", "where to write output JSON results (default stdout)")
	help        = flag.Bool("help", false, "prints this help and list of available analyses")
	analyses    = utils.CommaSeparatedFlags("analyses", "all", "comma-separated list of static analysis tasks to perform")
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

func checkTasks(names []string) ([]staticanalysis.Task, error) {
	uniqueNames := utils.RemoveDuplicates(names)
	var tasks []staticanalysis.Task
	for _, name := range uniqueNames {
		if task, ok := staticanalysis.TaskFromString(name); !ok {
			return nil, fmt.Errorf("unrecognised static analysis task '%s'", name)
		} else if task == staticanalysis.All {
			tasks = append(tasks, staticanalysis.AllTasks()...)
		} else {
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

func printAllTasks() {
	fmt.Fprintln(os.Stderr, "Available analysis tasks are:")
	for _, task := range staticanalysis.AllTasks() {
		fmt.Fprintln(os.Stderr, task)
	}
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
	startTime := time.Now()

	log.Initialize(os.Getenv("LOGGER_ENV"))

	flag.TextVar(&ecosystem, "ecosystem", pkgecosystem.None, pkgecosystem.FlagUsage+"(required)")
	analyses.InitFlag()
	flag.Parse()

	if len(os.Args) == 1 || *help == true {
		flag.Usage()
		fmt.Fprintln(os.Stderr, "")
		printAllTasks()
		return
	}

	if ecosystem == pkgecosystem.None || *packageName == "" {
		flag.Usage()
		return fmt.Errorf("ecosystem and package are required arguments")
	}

	manager := pkgmanager.Manager(ecosystem, false)
	if manager == nil {
		return fmt.Errorf("unsupported pkg manager for ecosystem %s", ecosystem)
	}

	pkg, err := worker.ResolvePkg(manager, *packageName, *version, *localFile)
	if err != nil {
		return fmt.Errorf("package error: %v", err)
	}

	analysisTasks, err := checkTasks(analyses.Values)
	if err != nil {
		printAllTasks()
		return err
	}

	log.Info("Static analysis launched",
		log.Label("ecosystem", ecosystem.String()),
		"package", *packageName,
		"version", *version,
		"local_path", *localFile,
		"output_file", *output,
		"analyses", analysisTasks)

	workDirs, err := makeWorkDirs()
	if err != nil {
		return fmt.Errorf("failed to create work directories: %v", err)
	}
	defer workDirs.cleanup()

	startExtractionTime := time.Now()
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

	extractionTime := time.Since(startExtractionTime)

	jsParserConfig, parserInitErr := parsing.InitParser(path.Join(workDirs.parserDir, jsParserDirName))
	if parserInitErr != nil {
		log.Error("failed to init JS parser", "error", parserInitErr)
	}

	startAnalysisTime := time.Now()
	results, err := staticanalysis.AnalyzePackageFiles(workDirs.extractDir, jsParserConfig, analysisTasks)
	analysisTime := time.Since(startAnalysisTime)
	if err != nil {
		return fmt.Errorf("static analysis error: %w", err)
	}

	startWritingResultsTime := time.Now()

	jsonResult, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("JSON marshall error: %v", err)
	}

	outputFile := os.Stdout
	if *output != "" {
		outputFile, err = os.Create(*output)
		if err != nil {
			return fmt.Errorf("could not open/create output file %s: %v", *output, err)
		}

		defer func() {
			if err := outputFile.Close(); err != nil {
				log.Warn("could not close output file", "path", *output, "error", err)
			}
		}()
	}

	if _, writeErr := outputFile.Write(jsonResult); writeErr != nil {
		return fmt.Errorf("could not write JSON results: %v", writeErr)
	}

	writingResultsTime := time.Since(startWritingResultsTime)

	totalTime := time.Since(startTime)
	otherTime := totalTime - writingResultsTime - analysisTime - extractionTime

	log.Info("Execution times",
		"download and extraction", extractionTime,
		"analysis", analysisTime,
		"writing results", writingResultsTime,
		"other", otherTime,
		"total", totalTime)

	return nil
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
