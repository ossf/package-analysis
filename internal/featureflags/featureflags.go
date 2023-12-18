package featureflags

import (
	"errors"
	"fmt"
	"strings"
)

var ErrUndefinedFlag = errors.New("undefined feature flag")

var flagRegistry = make(map[string]*FeatureFlag)

// FeatureFlag stores the state for a single flag.
//
// Call Enabled() to see if the flag is enabled.
type FeatureFlag struct {
	isEnabled bool
}

// new registers the flag and sets the default enabled state.
func new(name string, defaultEnabled bool) *FeatureFlag {
	ff := &FeatureFlag{
		isEnabled: defaultEnabled,
	}
	flagRegistry[name] = ff
	return ff
}

// Enabled returns whether or not the feature is enabled.
func (ff *FeatureFlag) Enabled() bool {
	return ff.isEnabled
}

// Update changes the internal state of the flags based on flags passed in.
//
// flags is a comma separated list of flag names. If a flag name is present it
// will be enabled. If a flag name is preceeded with a "-" character it will be
// disabled.
//
// For example: "MyFeature,-ExperimentalFeature" will enable the flag "MyFeature"
// and disable the flag "ExperimentalFeature".
//
// If a flag is undefined an error wrapping ErrUndefinedFlag will be returned.
func Update(flags string) error {
	if flags == "" {
		return nil
	}
	for _, n := range strings.Split(flags, ",") {
		isEnabled := true
		if n[0] == '-' {
			isEnabled = false
			n = n[1:]
		}
		if ff, ok := flagRegistry[n]; ok {
			ff.isEnabled = isEnabled
		} else {
			return fmt.Errorf("%w %q", ErrUndefinedFlag, n)
		}
	}
	return nil
}

// State returns a representation of the flags that are enabled and disabled.
func State() map[string]bool {
	s := make(map[string]bool)
	for k, v := range flagRegistry {
		s[k] = v.Enabled()
	}
	return s
}
