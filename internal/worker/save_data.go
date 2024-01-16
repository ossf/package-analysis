package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ossf/package-analysis/internal/featureflags"
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/resultstore"
	"github.com/ossf/package-analysis/internal/staticanalysis"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
	staticapi "github.com/ossf/package-analysis/pkg/api/staticanalysis"
)

// ResultStores holds ResultStore instances for saving each kind of analysis data.
// They can be nil, in which case calling the associated Upload function here is a no-op
type ResultStores struct {
	AnalyzedPackage      *resultstore.ResultStore
	DynamicAnalysis      *resultstore.ResultStore
	ExecutionLog         *resultstore.ResultStore
	FileWrites           *resultstore.ResultStore
	StaticAnalysis       *resultstore.ResultStore
	AnalyzedPackageSaved bool
}

// SaveDynamicAnalysisData saves the data from dynamic analysis to the corresponding bucket in the ResultStores.
// This includes strace data, execution log, and file writes (in that order).
// If any operation fails, the rest are aborted
func SaveDynamicAnalysisData(ctx context.Context, pkg *pkgmanager.Pkg, dest *ResultStores, data analysisrun.DynamicAnalysisData) error {
	if dest.DynamicAnalysis == nil {
		// nothing to do
		return nil
	}

	if err := dest.DynamicAnalysis.SaveDynamicAnalysis(ctx, pkg, data.StraceSummary, ""); err != nil {
		return fmt.Errorf("failed to save strace data to %s: %w", dest.DynamicAnalysis, err)
	}
	if err := saveExecutionLog(ctx, pkg, dest, data); err != nil {
		return err
	}
	if !featureflags.WriteFileContents.Enabled() {
		// Abort writing file contents when feature is disabled.
		return nil
	}
	if err := SaveFileWritesData(ctx, pkg, dest, data); err != nil {
		return err
	}
	if !featureflags.SaveAnalyzedPackages.Enabled() {
		// Abort saving analyzed packages when feature is disabled.
		return nil
	}
	if !dest.AnalyzedPackageSaved {
		if err := SaveAnalyzedPackage(ctx, pkg, dest); err != nil {
			return err
		} else {
			dest.AnalyzedPackageSaved = true
		}
	}

	return nil
}

// saveExecutionLog saves the execution log to the dynamic analysis resultstore, only if it is nonempty
func saveExecutionLog(ctx context.Context, pkg *pkgmanager.Pkg, dest *ResultStores, data analysisrun.DynamicAnalysisData) error {
	if dest.ExecutionLog == nil || len(data.ExecutionLog) == 0 {
		// nothing to do
		return nil
	}

	execLogFilename := "execution-log.json"
	if pkg.Version() != "" {
		execLogFilename = fmt.Sprintf("execution-log-%s.json", pkg.Version())
	}

	if err := dest.ExecutionLog.SaveDynamicAnalysis(ctx, pkg, data.ExecutionLog, execLogFilename); err != nil {
		return fmt.Errorf("failed to save execution log to %s: %w", dest.DynamicAnalysis, err)
	}

	return nil
}

// SaveStaticAnalysisData saves the data from static analysis to the corresponding bucket in the ResultStores
func SaveStaticAnalysisData(ctx context.Context, pkg *pkgmanager.Pkg, dest *ResultStores, data staticapi.SandboxData) error {
	if dest.StaticAnalysis == nil {
		return nil
	} else if len(data) == 0 {
		slog.WarnContext(ctx, "static analysis data is empty")
		return nil
	}

	var internalResult staticanalysis.Result
	if err := json.Unmarshal(data, &internalResult); err != nil {
		return fmt.Errorf("failed to unmarshal JSON data from sandbox into staticanalysis.Result: %w", err)
	}

	key := analysisrun.Key{
		Ecosystem: pkg.Ecosystem(),
		Name:      pkg.Name(),
		Version:   pkg.Version(),
	}
	serializableResult := internalResult.ToAPIResults()
	record := staticapi.CreateRecord(serializableResult, key)

	if err := dest.StaticAnalysis.SaveStaticAnalysis(ctx, pkg, record, ""); err != nil {
		return fmt.Errorf("failed to save static analysis results to %s: %w", dest.StaticAnalysis, err)
	}

	if !featureflags.SaveAnalyzedPackages.Enabled() {
		// Abort saving analyzed packages when feature is disabled.
		return nil
	}

	if !dest.AnalyzedPackageSaved {
		if err := SaveAnalyzedPackage(ctx, pkg, dest); err != nil {
			return err
		} else {
			dest.AnalyzedPackageSaved = true
		}
	}

	return nil
}

// SaveFileWritesData saves file writes data from dynamic analysis to the file writes bucket in the ResultStores
func SaveFileWritesData(ctx context.Context, pkg *pkgmanager.Pkg, dest *ResultStores, data analysisrun.DynamicAnalysisData) error {
	if dest.FileWrites == nil {
		return nil
	}

	fileWriteDataUploadStart := time.Now()
	if err := saveFileWriteResults(dest.FileWrites, ctx, pkg, data); err != nil {
		return fmt.Errorf("failed to save file write results to %s: %w", dest.FileWrites, err)
	}
	fileWriteDataDuration := time.Since(fileWriteDataUploadStart)

	slog.InfoContext(ctx, "Write data upload duration",
		"write_data_upload_duration", fileWriteDataDuration)

	return nil
}

// SaveAnalyzedPackage saves the analyzed package from static and dynamic analysis to the analyzed packages bucket in the ResultStores
func SaveAnalyzedPackage(ctx context.Context, pkg *pkgmanager.Pkg, dest *ResultStores) error {
	if pkg.IsLocal() {
		return nil
	}

	if err := dest.AnalyzedPackage.SaveAnalyzedPackage(ctx, pkg.Manager(), pkg); err != nil {
		return fmt.Errorf("failed to upload analyzed package to %s: %w", dest.AnalyzedPackage, err)
	}

	return nil
}
