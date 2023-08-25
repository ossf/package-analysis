package linelengths

import (
	"bufio"
	"io"
	"os"
	"strings"
)

/*
GetLineLengths counts the number of characters on each line of a file or string,
returning a slice containing the length of each line in sequence.

Lines are defined to be separated by newline ('\n') characters. If the newline
character is preceded by a carriage return ('\r'), this will also be treated as
part of the separator.

If filePath is not empty, the function attempts to count the lines of the file
at that path, otherwise lines in sourceString are counted.

Note: there may not be much useful information to be gathered by distinguishing
between line lengths when they get very long. It may be pragmatic to just report
all lines above e.g. 64K as 64K long.
*/
func GetLineLengths(filePath string, sourceString string) ([]int, error) {
	var reader *bufio.Reader
	if len(filePath) > 0 {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		reader = bufio.NewReader(file)
	} else {
		reader = bufio.NewReader(strings.NewReader(sourceString))
	}

	lengths := make([]int, 0)
	for {
		/* Normally bufio.Scanner would be more convenient to use here, however by default
		it uses a fixed maximum buffer size (MaxScanTokenSize = 64 * 1024). Since some
		(obfuscated) source code may contain very long lines, rather than doing our own
		buffer management we'll use reader.ReadStrings, which uses an internal function
		(collectFragments) to aggregate multiple full buffers. */
		line, readErr := reader.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return nil, readErr
		}

		// remove trailing newline and carriage return if present
		// (code adapted from bufio.ReadLine())
		l := len(line)
		if l >= 1 {
			if line[l-1] == '\n' {
				drop := 1
				if l >= 2 && line[l-2] == '\r' {
					drop = 2
				}
				l -= drop
			}
			lengths = append(lengths, l)
		}

		if readErr == io.EOF {
			break
		}
	}

	if len(lengths) == 0 {
		// define the empty string to have a single empty line
		lengths = append(lengths, 0)
	}

	return lengths, nil
}
