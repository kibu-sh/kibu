package ctxutil

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewContextLoader(t *testing.T) {
	key := struct{}{}
	type User struct {
		ID string
	}

	store := NewStore[User](key)

	t.Run("should be able to save and load a value from a context", func(t *testing.T) {
		expected := &User{ID: "test"}
		ctx := store.Save(context.Background(), expected)
		actual, err := store.Load(ctx)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	})

	t.Run("should return an error when the value is not found in the context", func(t *testing.T) {
		_, err := store.Load(context.Background())
		require.ErrorIs(t, err, ErrNotFoundInContext)
	})

	t.Run("should return an error with text context", func(t *testing.T) {
		_, err := store.Load(context.Background())
		require.Contains(t, err.Error(), "*ctxutil.User")
	})
}
