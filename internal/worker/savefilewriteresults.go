package worker

import (
	"context"
	"errors"
	"fmt"

	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/resultstore"
	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
)

func saveFileWriteResults(rs *resultstore.ResultStore, ctx context.Context, pkg *pkgmanager.Pkg, data analysisrun.DynamicAnalysisData) error {
	if rs == nil {
		// TODO this should become a method on resultstore.ResultStore?
		return errors.New("resultstore is nil")
	}

	if err := rs.SaveDynamicAnalysis(ctx, pkg, data.FileWritesSummary, ""); err != nil {
		return fmt.Errorf("failed to upload file write analysis to blobstore = %w", err)
	}
	var allPhasesWriteBufferIdsArray []string
	for _, writeBufferIds := range data.FileWriteBufferIds {
		allPhasesWriteBufferIdsArray = append(allPhasesWriteBufferIdsArray, writeBufferIds...)
	}

	// Remove potential duplicates across phases.
	allPhasesWriteBufferIdsArray = utils.RemoveDuplicates(allPhasesWriteBufferIdsArray)
	version := pkg.Version()
	if err := rs.SaveTempFilesToZip(ctx, pkg, "write_buffers_"+version, allPhasesWriteBufferIdsArray); err != nil {
		return fmt.Errorf("failed to upload file write buffer results to blobstore = #{err}")
	}
	if err := utils.RemoveTempFilesDirectory(); err != nil {
		return fmt.Errorf("failed to remove temp files = #{err}")
	}
	return nil
}
