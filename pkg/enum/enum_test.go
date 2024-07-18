package enum

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type isoStateCode3 string

func (s isoStateCode3) Validate() error {
	return isoStateSet.Validate(s)
}

var (
	isoStateSet = NewSet[isoStateCode3]()
	alabama     = isoStateSet.Add("AL", "Alabama")
	arizona     = isoStateSet.Add("AZ", "Arizona")
	arkansas    = isoStateSet.Add("AR", "Arkansas")
	california  = isoStateSet.Add("CA", "California")
)

func TestEnum(t *testing.T) {
	type testCase struct {
		input    isoStateCode3
		expected isoStateCode3
		err      error
		assert   func(t *testing.T, tc testCase)
	}

	var testCases = map[string]testCase{
		"should return the expected value": {
			input:    "AZ",
			expected: arizona,
		},
		"should return correct item for AL": {
			input:    "AL",
			expected: alabama,
		},
		"should return correct item for CA": {
			input:    "CA",
			expected: california,
		},
		"should return correct item for AR": {
			input:    "AR",
			expected: arkansas,
		},
		"should return an error when the value is not in the set": {
			input: "TX",
			err:   ErrInvalidType,
		},
		"should fail validation": {
			input: "TX",
			err:   ErrInvalidType,
			assert: func(t *testing.T, tc testCase) {
				require.ErrorIs(t, isoStateSet.Validate(tc.input), tc.err)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			item, err := isoStateSet.GetOrError(tc.input)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expected, item.ID)

			if tc.assert != nil {
				tc.assert(t, tc)
			}
		})
	}
}
