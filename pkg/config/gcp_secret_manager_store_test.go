package config

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGCPSecretManagerStore(t *testing.T) {
	ctx := context.Background()
	t.Skip("FIXME: skipping test until we can get a test account")

	temp, err := os.MkdirTemp("", "cloud_store_test")
	require.NoError(t, err)
	defer os.RemoveAll(temp)

	s, err := NewGCPSecretManagerStore(ctx, "TEST")
	require.NoError(t, err)
	relativeSecretPath := "level1/secret.json"

	var expected = map[string]any{
		"test": "test",
	}
	_, err = s.Set(ctx, SetParams{
		Path: relativeSecretPath,
		Data: expected,
	})
	require.NoError(t, err)

	var result map[string]any
	_, err = s.Get(ctx, GetParams{
		Path:   relativeSecretPath,
		Result: &result,
	})
	require.NoError(t, err)
	require.Equal(t, expected, result)
}
