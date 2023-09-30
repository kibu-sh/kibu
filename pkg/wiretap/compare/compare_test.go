package compare

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name          string
		value         string
		shouldContain string
		expected      bool
		expectedErr   error
	}{
		{
			name:          "string should contain hello",
			value:         "hello world",
			shouldContain: "hello",
			expected:      true,
		},
		{
			name:          "string should not contain hello",
			value:         "world",
			shouldContain: "hello",
			expected:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := Contains(test.shouldContain)(test.value)
			if test.expectedErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equalf(t, test.expected, actual, "expected %v to contain %v", test.value, test.shouldContain)
		})
	}
}

func TestExactly(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		exact     string
		expected  bool
		expectErr bool
	}{
		{
			name:     "string should match",
			value:    "hello",
			exact:    "hello",
			expected: true,
		},
		{
			name:     "string should not match",
			value:    "world",
			exact:    "hello",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := Exactly(test.exact)(test.value)
			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equalf(t, test.expected, actual, "expected %v to contain %v", test.value, test.exact)
		})
	}
}

func TestGlob(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		patten    string
		expected  bool
		expectErr bool
	}{
		{
			name:     "string should contain hello",
			value:    "hello world",
			patten:   "hello*",
			expected: true,
		},
		{
			name:     "string should not contain hello",
			value:    "world",
			patten:   "hello*",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := Glob(test.patten)(test.value)
			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equalf(t, test.expected, actual, "expected %v to contain %v", test.value, test.patten)
		})
	}

}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		prefix    string
		expected  bool
		expectErr bool
	}{
		{
			name:     "string should contain hello",
			value:    "hello world",
			prefix:   "hello",
			expected: true,
		},
		{
			name:     "string should not contain hello",
			value:    "world",
			prefix:   "hello",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := HasPrefix(test.prefix)(test.value)
			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equalf(t, test.expected, actual, "expected %v to contain %v", test.value, test.prefix)
		})
	}
}

func TestJSONPath(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		body          string
		shouldContain any
		expected      bool
		expectedErr   error
	}{
		{
			name:          "body should contain hello",
			path:          "hello",
			shouldContain: "world",
			body:          `{"hello": "world"}`,
			expected:      true,
		},
		{
			name:          "should handle malformed JSON",
			path:          "hello",
			shouldContain: "world",
			body:          `"world"}`,
			expected:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := strings.NewReader(test.body)
			actual, err := JSON(test.path, Contains(test.shouldContain.(string)))(body)
			if test.expectedErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equalf(t, test.expected, actual, "expected %v to contain %v", test.path, test.shouldContain)
		})
	}
}
