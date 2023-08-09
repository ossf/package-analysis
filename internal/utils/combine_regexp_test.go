package utils

import (
	"reflect"
	"regexp"
	"testing"
)

type combineRegexpTestCase struct {
	name    string
	regexps []*regexp.Regexp
	want    *regexp.Regexp
}

func TestCombineRegexp(t *testing.T) {
	tests := []combineRegexpTestCase{
		{
			name: "a b c",
			regexps: []*regexp.Regexp{
				regexp.MustCompile("a"),
				regexp.MustCompile("b"),
				regexp.MustCompile("c"),
			},
			want: regexp.MustCompile("(?:a)|(?:b)|(?:c)"),
		},
		{
			name: "capturing groups",
			regexps: []*regexp.Regexp{
				regexp.MustCompile("([0-9])"),
				regexp.MustCompile("([a-z])"),
				regexp.MustCompile("([A-Z])"),
			},
			want: regexp.MustCompile("(?:([0-9]))|(?:([a-z]))|(?:([A-Z]))"),
		},
		{
			name: "conjunction and capturing groups",
			regexps: []*regexp.Regexp{
				regexp.MustCompile("(apple|pear)"),
				regexp.MustCompile("(red|blue)"),
				regexp.MustCompile("(up|down)"),
			},
			want: regexp.MustCompile("(?:(apple|pear))|(?:(red|blue))|(?:(up|down))"),
		},
		{
			name: "quantification",
			regexps: []*regexp.Regexp{
				regexp.MustCompile("[!@#$%^&*()]{1, 30}"),
				regexp.MustCompile("\\s+"),
				regexp.MustCompile("[[:xdigit:]]?"),
			},
			want: regexp.MustCompile("(?:[!@#$%^&*()]{1, 30})|(?:\\s+)|(?:[[:xdigit:]]?)"),
		},
		{
			name: "combine regexps with quantifications",
			regexps: []*regexp.Regexp{
				regexp.MustCompile("(apple|pear)"),
				regexp.MustCompile("(red|blue)"),
				regexp.MustCompile("(up|down)"),
			},
			want: regexp.MustCompile("(?:(apple|pear))|(?:(red|blue))|(?:(up|down))"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CombineRegexp(tt.regexps...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CombineRegexp() = %v, want %v", got, tt.want)
			}
		})
	}
}
