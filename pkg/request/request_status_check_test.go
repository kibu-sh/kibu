package request

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStatusCheckFunc(t *testing.T) {
	tests := map[string]struct {
		status    string
		code      int
		expectErr error
		checkFunc StatusCheckFunc
	}{
		"should return nil if status is 200": {
			code:      200,
			checkFunc: NewOkayRangeCheckFunc(),
		},
		"should return nil if status is 201": {
			code:      201,
			checkFunc: NewOkayRangeCheckFunc(),
		},
		"should error if status is 300": {
			code:      300,
			checkFunc: NewOkayRangeCheckFunc(),
			expectErr: ErrStatusCheckFailed,
		},
		"should error if status is > 300": {
			code:      400,
			checkFunc: NewOkayRangeCheckFunc(),
			expectErr: ErrStatusCheckFailed,
		},
		"should error if status is not exactly 200": {
			code:      201,
			checkFunc: NewBasicOkayCheckFunc(),
			expectErr: ErrStatusCheckFailed,
		},
		"should return nil if status is exactly 200": {
			code:      200,
			checkFunc: NewBasicOkayCheckFunc(),
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			err := testCase.checkFunc(testCase.status, testCase.code)
			if testCase.expectErr != nil {
				require.ErrorIs(t, err, testCase.expectErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
