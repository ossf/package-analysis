package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/staticanalysis"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/worker"
	"github.com/ossf/package-analysis/pkg/api"
)

var (
	ecosystem   = flag.String("ecosystem", "", "Package ecosystem (required)")
	packageName = flag.String("package", "", "Package name (required)")
	version     = flag.String("version", "", "Package version (ignored if local file is specified)")
	localFile   = flag.String("local", "", "Local package archive containing package to be analysed. Name must match -package argument")
	output      = flag.String("output", "", "where to write output JSON results (default stdout)")
	help        = flag.Bool("help", false, "Prints this help and list of available analyses")
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

func doStaticAnalysis(d workDirs) (*staticanalysis.Result, error) {
	jsParserConfig, parserInitErr := parsing.InitParser(path.Join(d.parserDir, jsParserDirName))
	if parserInitErr != nil {
		log.Error("failed to init JS parser", "error", parserInitErr)
	}

	result := &staticanalysis.Result{
		BasicData:   map[string]staticanalysis.BasicFileData{},
		ParseData:   map[string]parsing.Result{},
		Obfuscation: nil,
	}

	walkErr := filepath.WalkDir(d.extractDir, func(path string, f fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if f.Type().IsRegular() {
			pathInArchive := strings.TrimPrefix(path, d.extractDir+string(os.PathSeparator))
			log.Info("Processing " + pathInArchive)

			result.BasicData[pathInArchive] = staticanalysis.GetBasicFileData(path, pathInArchive)

			parseData, parseErr := parsing.ParseSingle(jsParserConfig, path, "", false)
			if parseErr != nil {
				log.Error("Error parsing file", "filename", pathInArchive, "error", parseErr)
				result.ParseData[pathInArchive] = nil
			} else {
				result.ParseData[pathInArchive] = parseData
			}
		}
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("error while walking package files: %w", walkErr)
	}

	result.Obfuscation = obfuscation.ComputeResult(result.ParseData)

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
	startTime := time.Now()

	log.Initialize(os.Getenv("LOGGER_ENV"))
	flag.Parse()

	if len(os.Args) == 1 || *help == true {
		flag.Usage()
		return
	}

	if *ecosystem == "" || *packageName == "" {
		flag.Usage()
		return fmt.Errorf("ecosystem and package are required arguments")
	}

	manager := pkgecosystem.Manager(api.Ecosystem(*ecosystem))
	if manager == nil {
		return fmt.Errorf("unsupported pkg manager for ecosystem %s", *ecosystem)
	}

	pkg, err := worker.ResolvePkg(manager, *packageName, *version, *localFile)
	if err != nil {
		return fmt.Errorf("package error: %v", err)
	}

	log.Info("Static analysis launched",
		log.Label("ecosystem", *ecosystem),
		"package", *packageName,
		"version", *version,
		"local_path", *localFile,
		"output_file", *output)

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

	startAnalysisTime := time.Now()
	results, err := doStaticAnalysis(workDirs)
	analysisTime := time.Since(startAnalysisTime)

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
