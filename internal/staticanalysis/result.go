package staticanalysis

/*
Result (staticanalysis.Result) is the top-level data structure that stores
static analysis results. It holds combined data from each individual static
analysis task (see Task) performed on a package / artifact. Note that this
data is sent across a sandbox boundary, so must be serialisable.
*/
type Result map[Task]any
