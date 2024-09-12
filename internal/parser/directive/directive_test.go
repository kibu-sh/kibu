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
	{"kibu:service", true},
	{"kibu:worker", true},
	{"kibu:endpoint", true},
	{"kibu:workflow", true},
	{"kibu:activity", true},
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
			in:  "kibu:service",
			out: Directive{Tool: "kibu", Name: "service", Options: NewOptionList()},
		},

		"should parse directive with options": {
			in: "kibu:endpoint method=GET path=/thing/place/:location",
			out: Directive{
				Tool: "kibu", Name: "endpoint",
				Options: NewOptionListWithDefaults(map[string][]string{
					"method": []string{"GET"}, "path": []string{"/thing/place/:location"},
				}),
			},
		},

		"should handle misc spaces": {
			in: "kibu:endpoint method=GET  path=/thing/place/:location",
			out: Directive{
				Tool: "kibu", Name: "endpoint",
				Options: NewOptionListWithDefaults(map[string][]string{
					"method": []string{"GET"}, "path": []string{"/thing/place/:location"},
				}),
			},
		},
		"should parse multiple options": {
			in: "kibu:endpoint method=GET method=POST",
			out: Directive{
				Tool: "kibu", Name: "endpoint",
				Options: NewOptionListWithDefaults(map[string][]string{
					"method": []string{"GET", "POST"},
				}),
			},
		},

		"should parse multiple options with comma": {
			in: "kibu:endpoint method=GET,POST",
			out: Directive{
				Tool: "kibu", Name: "endpoint",
				Options: NewOptionListWithDefaults(map[string][]string{
					"method": []string{"GET", "POST"},
				}),
			},
		},

		"should error when directive key cannot be parsed": {
			in:  "kibuendpoint",
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
