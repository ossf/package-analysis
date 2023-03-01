package utils

import (
	"flag"
	"strings"
)

// CommaSeparatedFlags creates a struct which can be used with the Golang flag library,
// to allow passing a comma-separated list of strings as a single command-line argument.
//
// Make sure to call InitFlag() on the returned struct before calling flag.Parse().
func CommaSeparatedFlags(name string, values []string, usage string) CommaSeparatedFlagsData {
	return CommaSeparatedFlagsData{
		Name:   name,
		Values: values,
		Info:   usage,
	}
}

type CommaSeparatedFlagsData struct {
	Name   string
	Values []string
	Info   string
}

func (csl *CommaSeparatedFlagsData) Set(values string) error {
	csl.Values = strings.Split(values, ",")
	return nil
}

func (csl *CommaSeparatedFlagsData) String() string {
	if csl.Values == nil {
		return ""
	} else {
		return strings.Join(csl.Values, ",")
	}
}

func (csl *CommaSeparatedFlagsData) InitFlag() {
	flag.Var(csl, csl.Name, csl.Info)
}
