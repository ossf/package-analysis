package staticanalysis

// A Task (static analysis task) refers to a particular type of static analysis to be performed.
type Task string

const (
	ObfuscationDetection Task = "obfuscation"
)

func AllTasks() []Task {
	return []Task{
		ObfuscationDetection,
	}
}

func TaskFromString(s string) (Task, bool) {
	switch Task(s) {
	case ObfuscationDetection:
		return ObfuscationDetection, true
	default:
		return "", false
	}
}
