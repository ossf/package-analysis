package detection

import "github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stats"

type Signals struct {
	StringLengthSummary       stats.SampleStatistics
	StringEntropySummary      stats.SampleStatistics
	CombinedStringEntropy     float64
	IdentifierLengthSummary   stats.SampleStatistics
	IdentifierEntropySummary  stats.SampleStatistics
	CombinedIdentifierEntropy float64
}
