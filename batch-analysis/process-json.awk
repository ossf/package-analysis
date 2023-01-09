# This script reads in a json file and converts two particular JSON structs to inline
# The first struct has the format
# {
#     "Name": "...",
#     "Type": "..."
# }
#
# and the second has the format
# {
#     "Value": ...,
#     "Raw": "..."
# }
#
# These structures are respectively converted to:
# { "Name": "...", "Type": "..." }
#
# and
#
# { "Value": "...", "Raw": "..." }
#
# while preserving leading indendation
#
#

# We want to maintain two DFAs / state machines.
# State machine 1:
#     /{/ -> /"Name": "<name>",/ -> /"Type": "<type>"/ -> /},?/
# State machine 2:
#     /{/ -> /"Value": "<value>",/ -> /"Raw": "<raw>"/ -> /},?/

BEGIN {
	FS = "\n"    # the entire line is a single field

	# line history
	current = "" # current line
	prev1 = ""   # previous line
	prev2 = ""   # 2nd previous line
	prev3 = ""   # 3nd previous line

	state1Pos = 0  # tracks position in state machine 1
	state2Pos = 0  # tracks position in state machine 2

	unprintedLines = 0 # how many lines have been buffered during a pending path traversal

	defaultCase = 1  # set to 0 by any of the regex match cases
}

{
	# update line history
	prev3 = prev2
	prev2 = prev1
	prev1 = current
	current = $0

	# reset default case state for current line
	defaultCase = 1
}

# Since AWK can only iterate forwards over lines, we have to buffer
# the input lines and delay printing of each line of input until
# we can determine whether it's part of a matching struct or not.
# If not, this function is used to flush the buffered lines of input.
# Since the structs are 4 lines long in the input, the buffer contains
# 3 previous lines.
function flushUnprintedLines() {
	if (unprintedLines >= 3) print prev3
	if (unprintedLines >= 2) print prev2
	if (unprintedLines >= 1) print prev1

	print current

	unprintedLines = 0



/^[ \t]*{/ {
	# print "case 1"
	defaultCase = 0
	if (state1Pos == 0) {
		state1Pos = 1 # advance
		# print "state1Pos = 1"
		unprintedLines = 1
	} else {
		state1Pos = 0 # reset
	}

	if (state2Pos == 0) {
		state2Pos = 1 # advance
		# print "state2Pos = 1"
		unprintedLines = 1
	} else {
		state2Pos = 0 # reset
	}

	if (state1Pos == 0 && state2Pos == 0) {
		flushUnprintedLines()
	}
}


/^[ \t]*"Name": ".*",$/ {
	# print "case 2"
	defaultCase = 0
	if (state1Pos == 1) {
		state1Pos = 2 # advance
		# print "state1Pos = 2"
		unprintedLines = 2
	} else {
		state1Pos = 0 # reset
	}

	state2Pos = 0

	if (state1Pos == 0 && state2Pos == 0) {
		flushUnprintedLines()
	}
}


/^[ \t]*"Type": ".*"$/ {
	# print "case 3"
	defaultCase = 0
	if (state1Pos == 2) {
		state1Pos = 3 # advance
		# print "state1Pos = 3"
		unprintedLines = 3
	} else {
		state1Pos = 0 # reset
	}

	state2Pos = 0

	if (state1Pos == 0 && state2Pos == 0) {
		flushUnprintedLines()
	}
}

/^[ \t]*"Value": .*,$/ {
	# print "case 4"
	defaultCase = 0
	if (state2Pos == 1) {
		state2Pos = 2 # advance
		# print "state2Pos = 2"
		unprintedLines = 2
	} else {
		state2Pos = 0 # reset
	}

	state1Pos = 0

	if (state1Pos == 0 && state2Pos == 0) {
		flushUnprintedLines()
	}
}


/^[ \t]*"Raw": ".*"$/ {
	# print "case 5"
	defaultCase = 0
	if (state2Pos == 2) {
		state2Pos = 3 # advance
		# print "path3State = 3"
		unprintedLines = 3
	} else {
		state2Pos = 0 # reset
	}

	state1Pos = 0

	if (state1Pos == 0 && state2Pos == 0) {
		flushUnprintedLines()
	}
}

/^[ \t]*},?/ {
	# print "case 6"
	defaultCase = 0

	if (state1Pos == 3 || state2Pos == 3) {
		# completed path, print last 4 lines joined together with leading whitespace
		whitespace=sprintf("%s", current)
		gsub("[},]", "", whitespace)
		combined=sprintf("%s %s %s %s", prev3, prev2, prev1, current)
		gsub("[ \t\n]+", " ", combined)
		printf "%s%s\n", whitespace, combined
		unprintedLines = 0
	} else {
		flushUnprintedLines()
	}

	# unconditionally reset since we can't be in any other state now
	state1Pos = 0
	state2Pos = 0
}

defaultCase ~ /1/ {
	# print "default case"
	state1Pos = 0
	# print "state1Pos = 0"
	state2Pos = 0
	# print "state2Pos = 0"
	flushUnprintedLines()
}
