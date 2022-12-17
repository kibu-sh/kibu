package repo

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOperation_String(t *testing.T) {
	for i := 0; i < int(OpEnd); i++ {
		_, ok := operationNames[Operation(i)]
		require.NotContainsf(t, "UNKNOWN", Operation(i).String(), "operation missing from operationNames: %d", i)
		require.True(t, ok, "operation %d is not defined in name map", i)
	}
}
