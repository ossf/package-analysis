package staticanalysis

// A Task (static analysis task) refers to a particular type of static analysis to be performed.
// Some tasks may depend on other tasks, for example Obfuscation depends on Parsing.
type Task string

// NOTE: the string values below should match the JSON field names in result.go.
const (
	// Basic analysis consists of information about a file that can be determined
	// without parsing, for example file size, file type and hash.
	Basic Task = "basic"

	// Parsing analysis involves using a programming language parser to extract
	// source code information from the file.
	Parsing Task = "parsing"

	// Obfuscation analysis involves using certain rules to detect the presence of
	// obfuscated code. It depends on the output of the Parsing task, but does not
	// require reading files directly.
	Obfuscation Task = "obfuscation"

	// All is not a task itself, but represents/'depends on' all other tasks.
	All Task = "all"
)

var allTasks = []Task{
	Basic,
	Parsing,
	Obfuscation,
	All,
}

func AllTasks() []Task {
	return allTasks[:]
}

func TaskFromString(s string) (Task, bool) {
	switch Task(s) {
	case Basic:
		return Basic, true
	case Parsing:
		return Parsing, true
	case Obfuscation:
		return Obfuscation, true
	case All:
		return All, true
	default:
		return "", false
	}
}
