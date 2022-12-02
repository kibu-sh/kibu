package config

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestFileStore(t *testing.T) {
	ctx := context.Background()

	key := EncryptionKey{
		Engine: "base64key",
		Key:    "smGbjm71Nxd1Ig5FS0wj9SlbzAIrnolCz9bQQ6uAhl4=", // 32 byte key
	}

	temp, err := os.MkdirTemp("", "file_store_test")
	require.NoError(t, err)
	defer os.RemoveAll(temp)

	s := NewDefaultFileStore(temp)
	relativeSecretPath := "level1/secret.json"

	var expected = map[string]any{
		"test": "test",
	}
	_, err = s.Set(ctx, SetParams{
		EncryptionKey: key,
		Path:          relativeSecretPath,
		Data:          expected,
	})
	require.NoError(t, err)
	require.FileExistsf(t, filepath.Join(temp, relativeSecretPath), "file should exist relative to config store root")

	var result map[string]any
	_, err = s.Get(ctx, GetParams{
		Path:   relativeSecretPath,
		Result: &result,
	})
	require.NoError(t, err)
	require.Equal(t, expected, result)
}
