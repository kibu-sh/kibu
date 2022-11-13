package directive

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var isDirectiveTests = []struct {
	in string
	ok bool
}{
	{"abc", false},
	{"go:inline", true},
	{"Go:inline", false},
	{"go:Inline", false},
	{":inline", false},
	{"lint:ignore", true},
	{"lint:1234", true},
	{"1234:lint", true},
	{"go: inline", false},
	{"go:", false},
	{"go:*", false},
	{"go:x*", true},
	{"export foo", true},
	{"extern foo", true},
	{"expert foo", false},
	{"devx:service", true},
	{"devx:worker", true},
	{"devx:endpoint", true},
	{"devx:workflow", true},
	{"devx:activity", true},
}

func TestIsDirective(t *testing.T) {
	for _, tt := range isDirectiveTests {
		if ok := IsDirective(tt.in); ok != tt.ok {
			t.Errorf("IsDirective(%q) = %v, want %v", tt.in, ok, tt.ok)
		}
	}
}

var parseDirectiveTests = map[string]struct {
	in  string
	out Directive
	err error
}{
	"should parse basic directive": {
		in:  "devx:service",
		out: Directive{Tool: "devx", Name: "service"},
	},

	"should parse directive with options": {
		in: "devx:endpoint method=GET path=/thing/place/:location",
		out: Directive{
			Tool: "devx", Name: "endpoint",
			Options: OptionList{"method": "GET", "path": "/thing/place/:location"},
		},
	},

	"should handle misc spaces": {
		in: "devx:endpoint method=GET  path=/thing/place/:location",
		out: Directive{
			Tool: "devx", Name: "endpoint",
			Options: OptionList{"method": "GET", "path": "/thing/place/:location"},
		},
	},
}

func TestParseDirective(t *testing.T) {
	for name, tt := range parseDirectiveTests {
		t.Run(name, func(t *testing.T) {
			dir, err := Parse(tt.in)
			if tt.err != nil {
				require.NoError(t, err)
			}
			require.Equal(t, tt.out, dir)
		})
	}
}
