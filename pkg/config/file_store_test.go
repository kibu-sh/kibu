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

	// openssl rand -base64 32 | tr '+/' '-_' | pbcopy
	// 32 byte base64 url encoded key
	sourceKey := EncryptionKey{
		Engine: "base64key",
		Key:    "smGbjm71Nxd1Ig5FS0wj9SlbzAIrnolCz9bQQ6uAhl4=",
	}

	destKey := EncryptionKey{
		Engine: "base64key",
		Key:    "3vxfO8iIxHZqC7lMdMr4XAYRxV3-yevmUd9AWlYSBd4=",
	}

	store1, temp1, cleanup1, err := createTempStore()
	require.NoError(t, err)
	defer cleanup1()

	store2, temp2, cleanup2, err := createTempStore()
	require.NoError(t, err)
	defer cleanup2()

	relativeSecretPath := "level1/secret.json"

	var expected = map[string]any{
		"test": "test",
		"complex": map[string]any{
			"structure": float64(1),
		},
	}

	t.Run("should write secret to file store", func(t *testing.T) {
		_, err = store1.Set(ctx, SetParams{
			EncryptionKey: sourceKey,
			Path:          relativeSecretPath,
			Data:          expected,
		})
		require.NoError(t, err)
		require.FileExistsf(t, filepath.Join(temp1, relativeSecretPath), "file should exist relative to config store root")
	})

	t.Run("should read secret from file store", func(t *testing.T) {
		var result map[string]any
		_, err = store1.Get(ctx, GetParams{
			Path:   relativeSecretPath,
			Result: &result,
		})
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("should list secret from file store", func(t *testing.T) {
		iter, err := store1.List(ctx, ListParams{
			Path: "level1",
		})
		require.NoError(t, err)

		for item := range iter.Next() {
			require.NoError(t, item.Error())
			require.Equal(t, relativeSecretPath, item.Path())
			var actual map[string]any
			_, err := item.Get(ctx, &actual)
			require.NoError(t, err)
			require.Equal(t, expected, actual)
		}
	})

	t.Run("should copy one secret from one store to another", func(t *testing.T) {
		err = CopyOne(ctx, CopyParams{
			SourcePath:     relativeSecretPath,
			Source:         store1,
			Destination:    store2,
			DestinationKey: destKey,
		})
		require.NoError(t, err)
		require.FileExistsf(t, filepath.Join(temp2, relativeSecretPath), "file should exist relative to config store root")
	})

	t.Run("should copy secrets recursively from one store to another", func(t *testing.T) {
		var tests = []struct {
			expected any
			file     string
		}{
			{
				expected: expected,
				file:     relativeSecretPath,
			},
			{
				expected: expected,
				file:     "level1/level2/secret.json",
			},
			{
				expected: expected,
				file:     "level1/level2/level3/secret.json",
			},
		}

		for _, test := range tests {
			_, err = store1.Set(ctx, SetParams{
				EncryptionKey: sourceKey,
				Path:          test.file,
				Data:          test.expected,
			})
			require.NoErrorf(t, err, "failed to write %s", test.file)
		}

		err = CopyRecursive(ctx, CopyParams{
			Source:         store1,
			Destination:    store2,
			DestinationKey: destKey,
		})
		require.NoError(t, err)

		for _, test := range tests {
			var actual map[string]any
			_, err = store2.Get(ctx, GetParams{
				Path:   test.file,
				Result: &actual,
			})
			require.NoErrorf(t, err, "failed to read %s", test.file)
			require.FileExistsf(t, filepath.Join(temp2, test.file), "file should exist relative to config store root")
			require.Equal(t, test.expected, actual)
		}
	})
}

func createTempStore() (store *FileStore, tempPath string, cleanup func(), err error) {
	tempPath, err = os.MkdirTemp("", "file_store_test")
	if err != nil {
		return
	}

	store = NewDefaultFileStore(tempPath)
	return store, tempPath, func() {
		_ = os.RemoveAll(tempPath)
	}, nil
}
