package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/staticanalysis"
	"github.com/ossf/package-analysis/internal/staticanalysis/basicdata"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/useragent"
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
	analyses    = utils.CommaSeparatedFlags("analyses", []string{"all"}, "comma-separated list of static analysis tasks to perform")
)

type workDirs struct {
	baseDir    string
	archiveDir string
	extractDir string
	parserDir  string
}

func (wd *workDirs) cleanup(ctx context.Context) {
	if err := os.RemoveAll(wd.baseDir); err != nil {
		slog.ErrorContext(ctx, "Failed to remove work dirs", "baseDir", wd.baseDir, "error", err)
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

	userAgentExtra := os.Getenv("OSSF_MALWARE_USER_AGENT_EXTRA")
	http.DefaultTransport = useragent.DefaultRoundTripper(http.DefaultTransport, userAgentExtra)

	flag.TextVar(&ecosystem, "ecosystem", pkgecosystem.None, fmt.Sprintf("package ecosystem. Can be %s (required)", pkgecosystem.SupportedEcosystemsStrings))
	analyses.InitFlag()
	flag.Parse()

	if len(os.Args) == 1 || *help {
		flag.Usage()
		fmt.Fprintln(os.Stderr, "")
		printAllTasks()
		return
	}

	if ecosystem == pkgecosystem.None || *packageName == "" {
		flag.Usage()
		return fmt.Errorf("ecosystem and package are required arguments")
	}

	manager := pkgmanager.Manager(ecosystem)
	if manager == nil {
		return fmt.Errorf("unsupported pkg manager for ecosystem %s", ecosystem)
	}

	pkg, err := worker.ResolvePkg(manager, *packageName, *version, *localFile)
	if err != nil {
		return fmt.Errorf("package error: %w", err)
	}

	analysisTasks, err := checkTasks(analyses.Values)
	if err != nil {
		printAllTasks()
		return err
	}

	ctx := log.ContextWithAttrs(context.Background(),
		log.Label("ecosystem", ecosystem.String()),
		slog.String("package", *packageName),
		slog.String("version", *version),
	)

	slog.InfoContext(ctx, "Static analysis launched",
		"local_path", *localFile,
		"output_file", *output,
		"analyses", analysisTasks,
		"user_agent_extra", userAgentExtra)

	workDirs, err := makeWorkDirs()
	if err != nil {
		return fmt.Errorf("failed to create work directories: %w", err)
	}
	defer workDirs.cleanup(ctx)

	startDownloadTime := time.Now()

	var archivePath string
	if *localFile != "" {
		archivePath = *localFile
	} else {
		archivePath, err = manager.DownloadArchive(pkg.Name(), pkg.Version(), workDirs.archiveDir)
		if err != nil {
			return fmt.Errorf("error downloading archive: %w", err)
		}
	}

	downloadTime := time.Since(startDownloadTime)

	results := staticanalysis.Result{}

	startArchiveAnalysisTime := time.Now()
	archiveResult, err := basicdata.Analyze(ctx, []string{archivePath},
		basicdata.SkipLineLengths(),
		basicdata.FormatPaths(func(absPath string) string { return "/" }),
	)
	if err != nil {
		slog.WarnContext(ctx, "failed to analyze archive file", "error", err)
	} else if len(archiveResult) != 1 {
		slog.WarnContext(ctx, "archive file analysis: unexpected number of results", "len", len(archiveResult))
	} else {
		archiveInfo := archiveResult[0]
		results.Archive = staticanalysis.ArchiveResult{
			DetectedType: archiveInfo.DetectedType,
			Size:         archiveInfo.Size,
			SHA256:       archiveInfo.SHA256,
		}
	}

	archiveAnalysisTime := time.Since(startArchiveAnalysisTime)

	startExtractionTime := time.Now()

	if err := manager.ExtractArchive(archivePath, workDirs.extractDir); err != nil {
		return fmt.Errorf("archive extraction failed: %w", err)
	}

	extractionTime := time.Since(startExtractionTime)

	jsParserConfig, parserInitErr := parsing.InitParser(ctx, filepath.Join(workDirs.parserDir, jsParserDirName))
	if parserInitErr != nil {
		slog.ErrorContext(ctx, "failed to init JS parser", "error", parserInitErr)
	}

	startAnalysisTime := time.Now()
	fileResults, err := staticanalysis.AnalyzePackageFiles(ctx, workDirs.extractDir, jsParserConfig, analysisTasks)
	if err != nil {
		return fmt.Errorf("static analysis error: %w", err)
	}
	results.Files = fileResults

	analysisTime := time.Since(startAnalysisTime)
	startWritingResultsTime := time.Now()

	jsonResult, err := json.Marshal(results)
	if err != nil {
		slog.WarnContext(ctx, fmt.Sprintf("unserialisable JSON: %v", results))
		return fmt.Errorf("JSON marshal error: %w", err)
	}

	outputFile := os.Stdout
	if *output != "" {
		outputFile, err = os.Create(*output)
		if err != nil {
			return fmt.Errorf("could not open/create output file %s: %w", *output, err)
		}

		defer func() {
			if err := outputFile.Close(); err != nil {
				slog.WarnContext(ctx, "could not close output file", "path", *output, "error", err)
			}
		}()
	}

	if _, writeErr := outputFile.Write(jsonResult); writeErr != nil {
		return fmt.Errorf("could not write JSON results: %w", writeErr)
	}

	writingResultsTime := time.Since(startWritingResultsTime)

	totalTime := time.Since(startTime)
	otherTime := totalTime - writingResultsTime - analysisTime - extractionTime - archiveAnalysisTime - downloadTime

	slog.InfoContext(ctx, "Execution times",
		"download", downloadTime,
		"archive analysis", archiveAnalysisTime,
		"archive extraction", extractionTime,
		"file analysis", analysisTime,
		"writing results", writingResultsTime,
		"other", otherTime,
		"total", totalTime)

	return nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("static analysis failed", "error", err)
		os.Exit(1)
	}
}
