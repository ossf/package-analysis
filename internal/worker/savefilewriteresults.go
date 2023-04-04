package worker

import (
	"context"
	"fmt"

	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/resultstore"
	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
)

func SaveFileWriteResults(bucket string, resultStoreOptions resultstore.Option, ctx context.Context, pkg *pkgmanager.Pkg, dynamicResults analysisrun.DynamicAnalysisResults) error {
	rs := resultstore.New(bucket, resultStoreOptions)
	if err := rs.Save(ctx, pkg, dynamicResults.FileWritesSummary); err != nil {
		return fmt.Errorf("failed to upload file write analysis to blobstore = %w", err)
	}
	var allPhasesWriteBufferIdsArray []string
	for _, writeBufferIds := range dynamicResults.FileWriteBufferIds {
		allPhasesWriteBufferIdsArray = append(allPhasesWriteBufferIdsArray, writeBufferIds...)
	}

	// Remove potential duplicates across phases.
	allPhasesWriteBufferIdsArray = utils.RemoveDuplicates(allPhasesWriteBufferIdsArray)

	if err := rs.SaveTempFilesToZip(ctx, pkg, "write_buffers", allPhasesWriteBufferIdsArray); err != nil {
		return fmt.Errorf("failed to upload file write buffer results to blobstore = #{err}")
	}
	if err := utils.RemoveTempFilesDirectory(); err != nil {
		return fmt.Errorf("failed to remove temp files = #{err}")
	}
	return nil
}
