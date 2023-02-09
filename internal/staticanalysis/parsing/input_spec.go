package parsing

import "fmt"

// InputSpec allows different ways of passing input to a parser,
// for example a list of files or a raw string
type InputSpec struct {
	filePaths []string
	rawString string
}

func (i InputSpec) isValid() error {
	if len(i.filePaths) > 0 && i.rawString != "" {
		return fmt.Errorf("invalid InputSpec: cannot both have file paths and raw string specified")
	}
	return nil
}

func MultipleFileInput(paths []string) InputSpec {
	return InputSpec{filePaths: paths}
}

func StringInput(sourceCode string) InputSpec {
	return InputSpec{rawString: sourceCode}
}
