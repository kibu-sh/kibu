package model

import (
	"github.com/discernhq/devx/pkg/database/xql"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReflectEntity(t *testing.T) {
	type User struct {
		ID      int    `db:"id,pk,table=users,schema=public"`
		Name    string `db:"name"`
		Ignored string `db:"-"`
	}
	type Related struct {
		ID int `db:"id,pk"`
	}
	t.Run("should infer table name from struct name", func(t *testing.T) {
		def, err := Reflect[Related]("db")
		require.NoError(t, err)
		require.Equal(t, "Related", def.table)
	})
	t.Run("should produce an model by reflecting a struct", func(t *testing.T) {
		def, err := Reflect[User]("db")
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

	t.Run("should produce correct primary key as a map", func(t *testing.T) {
		type UserWithCompositePK struct {
			ID   int    `db:"id,pk"`
			Name string `db:"name,pk"`
		}
		entity, err := Reflect[UserWithCompositePK]("db")
		require.NoError(t, err)
		require.Equal(t, xql.Eq{
			"id":   1,
			"name": "John",
		}, entity.PrimaryKeyPredicate(&UserWithCompositePK{
			ID:   1,
			Name: "John",
		}))
	})

	t.Run("should be able to map to and from model", func(t *testing.T) {
		type User struct {
			ID   int `db:"id"`
			Name string
		}
		user := &User{
			ID:   1,
			Name: "John",
		}
		def, err := Reflect[User]("db")
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
}
