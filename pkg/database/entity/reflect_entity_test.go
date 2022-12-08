package entity

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
		def, err := ReflectEntityDefinition[User, int]("db")
		require.NoError(t, err)
		require.Equal(t, "public", def.schema)
		require.Equal(t, "users", def.table)
		require.Equal(t, Fields{
			{Name: "id", IsIdentity: true},
			{Name: "name"},
		}, def.fields)

		require.Equal(t, map[string]structReflectMeta{
			"id": structReflectMeta{
				Name: "ID",
				ID:   0,
			},
			"name": structReflectMeta{
				Name: "Name",
				ID:   1,
			},
		}, def.dbToStruct)

		require.Equal(t, map[string]string{
			"ID":   "id",
			"Name": "name",
		}, def.structToDB)
	})

	t.Run("should produce correct primary key as a string", func(t *testing.T) {
		entity, err := ReflectEntityDefinition[User, int]("db")
		require.NoError(t, err)
		require.Equal(t, "id", entity.fields.PrimaryKey().String())
	})

	t.Run("should be able to map to and from entity", func(t *testing.T) {
		type User struct {
			ID   int `db:"id"`
			Name string
		}
		user := &User{
			ID:   1,
			Name: "John",
		}
		def, err := ReflectEntityDefinition[User, int]("db")
		require.NoError(t, err)

		valueMap := def.ValueMap(user)
		require.NoError(t, err)
		require.Equal(t, ValueMap{
			"id":   1,
			"Name": "John",
		}, valueMap)

		newUser := def.ValueMapToEntity(valueMap)
		require.Equal(t, user, newUser)

		values := def.ColumnValues(user)
		require.Equal(t, []any{1, "John"}, values)
	})

	t.Run("should support composite primary keys", func(t *testing.T) {})
}
