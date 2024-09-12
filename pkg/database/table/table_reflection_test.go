package table

import (
	"github.com/kibu-sh/kibu/pkg/database/xql"
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
	t.Run("should infer Table name from struct name", func(t *testing.T) {
		def, err := Reflect[Related](xql.SQLite3, "db")
		require.NoError(t, err)
		require.Equal(t, "Related", def.Table)
	})
	t.Run("should produce an Table by reflecting a struct", func(t *testing.T) {
		def, err := Reflect[User](xql.SQLite3, "db")
		require.NoError(t, err)
		require.Equal(t, "public", def.Schema)
		require.Equal(t, "users", def.Table)
		require.Equal(t, Columns{
			{Name: "id", IsIdentity: true},
			{Name: "name"},
		}, def.Columns)

		require.Equal(t, map[string]StructMetadata{
			"id": StructMetadata{
				Name: "ID",
				ID:   0,
			},
			"name": StructMetadata{
				Name: "Name",
				ID:   1,
			},
		}, def.DBToStruct)

		require.Equal(t, map[string]string{
			"ID":   "id",
			"Name": "name",
		}, def.StructToDB)
	})

	t.Run("should produce correct primary key as a map", func(t *testing.T) {
		type UserWithCompositePK struct {
			ID   int    `db:"id,pk"`
			Name string `db:"name,pk"`
		}
		entity, err := Reflect[UserWithCompositePK](xql.SQLite3, "db")
		require.NoError(t, err)
		require.Equal(t, xql.Eq{
			"id":   1,
			"name": "John",
		}, entity.PrimaryKeyPredicate(&UserWithCompositePK{
			ID:   1,
			Name: "John",
		}))
	})

	t.Run("should be able to map to and from Table", func(t *testing.T) {
		type User struct {
			ID   int `db:"id"`
			Name string
		}
		user := &User{
			ID:   1,
			Name: "John",
		}
		def, err := Reflect[User](xql.SQLite3, "db")
		require.NoError(t, err)

		valueMap := def.ValueMap(user)
		require.NoError(t, err)
		require.Equal(t, ValueMap{
			"id":   1,
			"Name": "John",
		}, valueMap)

		newUser := def.ValueMapToModel(valueMap)
		require.Equal(t, user, newUser)

		values := def.ColumnValues(user)
		require.Equal(t, []any{1, "John"}, values)
	})
}
