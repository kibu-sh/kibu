package config

import (
	"context"
	"github.com/stretchr/testify/require"
	"gocloud.dev/secrets"
	_ "gocloud.dev/secrets/localsecrets"
	"os"
	"path/filepath"
	"testing"
)

func TestFileStore(t *testing.T) {
	ctx := context.Background()
	key := "base64key://" // Using "base64key://", a new random key will be generated.

	temp, err := os.MkdirTemp("", "file_store_test")
	require.NoError(t, err)
	defer os.RemoveAll(temp)

	keeper, err := secrets.OpenKeeper(ctx, key)
	defer keeper.Close()
	require.NoError(t, err)

	s := &FileStore{
		key:     key,
		crypter: keeper,
		fs:      DirectoryFS{Path: temp},
	}

	var expected = map[string]any{
		"test": "test",
	}
	err = s.Set(ctx, SetParams{
		Key:  "test",
		Data: expected,
	})
	require.NoError(t, err)
	require.FileExistsf(t, filepath.Join(temp, "test"), "file should exist")

	var result map[string]any
	err = s.Get(ctx, GetParams{
		Key:    "test",
		Result: &result,
	})
	require.NoError(t, err)
	require.Equal(t, expected, result)
}
