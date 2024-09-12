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
	{"kibue:service", true},
	{"kibue:worker", true},
	{"kibue:endpoint", true},
	{"kibue:workflow", true},
	{"kibue:activity", true},
}

func TestIsDirective(t *testing.T) {
	for _, tt := range isDirectiveTests {
		if ok := IsDirective(tt.in); ok != tt.ok {
			t.Errorf("IsDirective(%q) = %v, want %v", tt.in, ok, tt.ok)
		}
	}
}

func TestParseDirective(t *testing.T) {
	var parseDirectiveTests = map[string]struct {
		in  string
		out Directive
		err error
	}{
		"should parse basic directive": {
			in:  "kibue:service",
			out: Directive{Tool: "kibue", Name: "service", Options: NewOptionList()},
		},

		"should parse directive with options": {
			in: "kibue:endpoint method=GET path=/thing/place/:location",
			out: Directive{
				Tool: "kibue", Name: "endpoint",
				Options: NewOptionListWithDefaults(map[string][]string{
					"method": []string{"GET"}, "path": []string{"/thing/place/:location"},
				}),
			},
		},

		"should handle misc spaces": {
			in: "kibue:endpoint method=GET  path=/thing/place/:location",
			out: Directive{
				Tool: "kibue", Name: "endpoint",
				Options: NewOptionListWithDefaults(map[string][]string{
					"method": []string{"GET"}, "path": []string{"/thing/place/:location"},
				}),
			},
		},
		"should parse multiple options": {
			in: "kibue:endpoint method=GET method=POST",
			out: Directive{
				Tool: "kibue", Name: "endpoint",
				Options: NewOptionListWithDefaults(map[string][]string{
					"method": []string{"GET", "POST"},
				}),
			},
		},

		"should parse multiple options with comma": {
			in: "kibue:endpoint method=GET,POST",
			out: Directive{
				Tool: "kibue", Name: "endpoint",
				Options: NewOptionListWithDefaults(map[string][]string{
					"method": []string{"GET", "POST"},
				}),
			},
		},

		"should error when directive key cannot be parsed": {
			in:  "kibueendpoint",
			err: ErrInvalidDirective,
		},

		// TODO: think about invalid options cases
		// How will we validate the option?
	}

	for name, tc := range parseDirectiveTests {
		t.Run(name, func(t *testing.T) {
			dir, err := Parse(tc.in)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.out, dir)
		})
	}
}
