package analysis

// Mode (analysis mode) is used to distinguish between whether static or dynamic analysis is being performed.
type Mode string

const (
	Dynamic Mode = "dynamic"
	Static  Mode = "static"
)

func AllModes() []Mode {
	return []Mode{Dynamic, Static}
}

func ModeFromString(s string) (Mode, bool) {
	switch Mode(s) {
	case Dynamic:
		return Dynamic, true
	case Static:
		return Static, true
	default:
		return "", false
	}
}
