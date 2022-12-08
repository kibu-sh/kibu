package database

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReflectEntity(t *testing.T) {
	type User struct {
		ID      int    `db:"id,pk,table=users,schema=public"`
		Name    string `db:"name"`
		Ignored string `db:"-"`
	}
	t.Run("should produce an entity by reflecting a struct", func(t *testing.T) {
		entity, err := ReflectEntityDefinition[User]("db")
		require.NoError(t, err)
		require.Equal(t, "public", entity.Schema)
		require.Equal(t, "users", entity.Table)
		require.Equal(t, Fields{
			{Name: "id", IsIdentity: true},
			{Name: "name"},
		}, entity.Fields)
	})

	t.Run("should produce correct primary key as a string", func(t *testing.T) {
		entity, err := ReflectEntityDefinition[User]("db")
		require.NoError(t, err)
		require.Equal(t, "id", entity.Fields.PrimaryKey().String())
	})

	t.Run("should support composite primary keys", func(t *testing.T) {})
}
