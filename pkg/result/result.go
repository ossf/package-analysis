package result

import (
	"github.com/ossf/package-analysis/internal/dynamicanalysis"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
)

type DynamicAnalysisStraceSummary map[pkgecosystem.RunPhase]*dynamicanalysis.StraceSummary
type DynamicAnalysisFileWrites map[pkgecosystem.RunPhase]*dynamicanalysis.FileWrites

type DynamicAnalysisResults struct {
	StraceSummary DynamicAnalysisStraceSummary
	FileWrites    DynamicAnalysisFileWrites
}
