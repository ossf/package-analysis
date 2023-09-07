package signals

// Result holds all data produced by signals analysis (see Analyze() in analyze.go).
type Result struct {
	// Files contains a signals.FileSignals object that is useful for detecting suspicious files.
	Files []FileSignals `json:"files"`
}
