package staticanalysis

// A Task (static analysis task) refers to a particular type of static analysis to be performed.
// Some tasks may depend on other tasks, for example Obfuscation depends on Parsing
type Task string

const (
	// BasicData analysis consists of information about a file that can be determined
	// without parsing, for example file size, file type and hash
	BasicData Task = "basic_data"

	// Parsing analysis involves using a programming language parser to extract
	// source code information from the file
	Parsing Task = "parsing"

	// Obfuscation analysis involves using certain rules to detect the presence of
	// obfuscated code. It depends on the output of the Parsing task, but does not
	// require reading files directly.
	Obfuscation Task = "obfuscation"
)

var allTasks = []Task{
	BasicData,
	Parsing,
	Obfuscation,
}

func AllTasks() []Task {
	return allTasks[:]
}

func TaskFromString(s string) (Task, bool) {
	switch Task(s) {
	case BasicData:
		return BasicData, true
	case Parsing:
		return Parsing, true
	case Obfuscation:
		return Obfuscation, true
	default:
		return "", false
	}
}
