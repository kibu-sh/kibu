package ctxutil

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

type key1 struct{}
type key2 struct{}

func TestNewContextLoader(t *testing.T) {

	type User struct {
		ID string
	}

	store := NewStore[User, key1]()
	store2 := NewStore[User, key2]()

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

	t.Run("keys should not collide", func(t *testing.T) {
		expected := &User{ID: "test"}
		isolated := context.Background()
		isolated = store.Save(isolated, expected)
		_, err := store2.Load(isolated)
		require.ErrorIs(t, err, ErrNotFoundInContext)

		got, err := store.Load(isolated)
		require.NoError(t, err)
		require.Equal(t, got.ID, expected.ID)
	})
}
