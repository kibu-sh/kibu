package database

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

type mockObject struct {
	ID   uuid.UUID
	Data map[string]string
}

func (m mockObject) PrimaryKey() string {
	return m.ID.String()
}

// FIXME
// compile check that MemoryRepository implements Repo
// var _ Repo[mockObject] = (*MemoryRepository[mockObject])(nil)

func TestMemRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository[mockObject]()

	t.Run("should create a new object", func(t *testing.T) {
		expected := &mockObject{
			ID: uuid.New(),
		}

		err := repo.Save(ctx, expected)
		require.NoError(t, err)

		model, err := repo.FindOneOrThrow(ctx, expected.PrimaryKey())
		require.NoError(t, err)
		require.Equal(t, expected, model)
	})

	t.Run("should delete an object", func(t *testing.T) {
		expected := &mockObject{
			ID: uuid.New(),
		}

		err := repo.Save(ctx, expected)
		require.NoError(t, err)

		err = repo.Delete(ctx, expected)
		require.NoError(t, err)

		model, err := repo.FindOne(ctx, expected.PrimaryKey())
		require.NoError(t, err)
		require.Nil(t, model)
	})

	t.Run("should throw an error when object not found", func(t *testing.T) {
		_, err := repo.FindOneOrThrow(ctx, uuid.New().String())
		require.ErrorIs(t, err, ErrNotFound)
	})

	// TODO: move to Update method
	t.Run("should update a table by its id", func(t *testing.T) {
		expected := &mockObject{
			ID:   uuid.New(),
			Data: make(map[string]string),
		}

		err := repo.Save(ctx, expected)
		require.NoError(t, err)

		expected.Data = map[string]string{
			"hello": "world",
		}
		err = repo.Save(ctx, expected)
		require.NoError(t, err)

		model, err := repo.FindOneOrThrow(ctx, expected.PrimaryKey())
		require.NoError(t, err)
		require.Equal(t, expected, model)
	})
}
