package analysis

// Mode is used to distinguish between whether static or dynamic analysis is being performed
type Mode string

const (
	Dynamic     Mode = "dynamic"
	Static      Mode = "static"
	InvalidMode Mode = ""
)

func AllModes() []Mode {
	return []Mode{Dynamic, Static}
}

func ModeFromString(s string) Mode {
	switch Mode(s) {
	case Dynamic:
		return Dynamic
	case Static:
		return Static
	default:
		return InvalidMode
	}
}
