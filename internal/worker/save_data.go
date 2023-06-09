package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/resultstore"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
)

// ResultStores holds ResultStore instances for saving each kind of analysis data.
// They can be nil, in which case calling the associated Upload function here is a no-op
type ResultStores struct {
	DynamicAnalysis *resultstore.ResultStore
	StaticAnalysis  *resultstore.ResultStore
	FileWrites      *resultstore.ResultStore
	AnalyzedPackage *resultstore.ResultStore
}

// SaveDynamicAnalysisData saves the data from dynamic analysis to the corresponding bucket in the ResultStores.
// This includes strace data, execution log, and file writes (in that order).
// If any operation fails, the rest are aborted
func SaveDynamicAnalysisData(ctx context.Context, pkg *pkgmanager.Pkg, dest ResultStores, data analysisrun.DynamicAnalysisResults) error {
	if dest.DynamicAnalysis == nil {
		// nothing to do
		return nil
	}

	if err := dest.DynamicAnalysis.Save(ctx, pkg, data.StraceSummary); err != nil {
		return fmt.Errorf("failed to save strace data to %s: %w", dest.DynamicAnalysis, err)
	}
	if err := saveExecutionLog(ctx, pkg, dest, data); err != nil {
		return err
	}
	if err := SaveFileWritesData(ctx, pkg, dest, data); err != nil {
		return err
	}

	if err := SaveAnalyzedPackage(ctx, pkg, dest); err != nil {
		return err
	}

	return nil
}

// saveExecutionLog saves the execution log to the dynamic analysis resultstore, only if it is nonempty
func saveExecutionLog(ctx context.Context, pkg *pkgmanager.Pkg, dest ResultStores, data analysisrun.DynamicAnalysisResults) error {
	if dest.DynamicAnalysis == nil || len(data.ExecutionLog) == 0 {
		// nothing to do
		return nil
	}

	execLogFilename := resultstore.MakeFilename(pkg, "execution-log")
	if err := dest.DynamicAnalysis.SaveWithFilename(ctx, pkg, execLogFilename, data.ExecutionLog); err != nil {
		return fmt.Errorf("failed to save execution log to %s: %w", dest.DynamicAnalysis, err)
	}

	return nil
}

// SaveStaticAnalysisData saves the data from static analysis to the corresponding bucket in the ResultStores
func SaveStaticAnalysisData(ctx context.Context, pkg *pkgmanager.Pkg, dest ResultStores, data analysisrun.StaticAnalysisResults) error {
	if dest.StaticAnalysis == nil {
		return nil
	}

	if err := dest.StaticAnalysis.Save(ctx, pkg, data); err != nil {
		return fmt.Errorf("failed to save static analysis results to %s: %w", dest.StaticAnalysis, err)
	}

	if err := SaveAnalyzedPackage(ctx, pkg, dest); err != nil {
		return err
	}

	return nil
}

// SaveFileWritesData saves file writes data from dynamic analysis to the file writes bucket in the ResultStores
func SaveFileWritesData(ctx context.Context, pkg *pkgmanager.Pkg, dest ResultStores, data analysisrun.DynamicAnalysisResults) error {
	if dest.FileWrites == nil {
		return nil
	}

	fileWriteDataUploadStart := time.Now()
	if err := saveFileWriteResults(dest.FileWrites, ctx, pkg, data); err != nil {
		return fmt.Errorf("failed to save file write results to %s: %w", dest.FileWrites, err)
	}
	fileWriteDataDuration := time.Since(fileWriteDataUploadStart)

	log.Info("Write data upload duration",
		log.Label("ecosystem", pkg.EcosystemName()),
		"name", pkg.Name(),
		"version", pkg.Version(),
		"write_data_upload_duration", fileWriteDataDuration,
	)

	return nil
}

// SaveAnalyzedPackage saves the analyzed package from static and dynamic analysis to the analyzed packages bucket in the ResultStores
func SaveAnalyzedPackage(ctx context.Context, pkg *pkgmanager.Pkg, dest ResultStores) error {
	if pkg.IsLocal() {
		return nil
	}

	if err := dest.AnalyzedPackage.SaveAnalyzedPackage(ctx, pkg.Manager(), pkg); err != nil {
		return fmt.Errorf("failed to upload analyzed package to blobstore = %w", err)
	}

	return nil
}