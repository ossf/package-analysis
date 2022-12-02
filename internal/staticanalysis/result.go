package staticanalysis

/*
Result (staticanalysis.Result) is the top-level static analysis result data structure.
It holds combined data from all of the individual static analysis tasks (see Task)
performed on a package / artifact. Note that this data must be sent across a sandbox
boundary, so it must be serialisable..
*/
type Result map[Task]any
