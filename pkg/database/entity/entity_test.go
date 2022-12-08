package entity

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEntityDefinition(t *testing.T) {
	t.Run("should produce a composite relation name", func(t *testing.T) {
		ed := Definition[any, any]{
			schema: "schema",
			table:  "table",
		}
		require.Equal(t, "schema.table", ed.RelationName())
	})

	t.Run("should exclude empty schema from relation name", func(t *testing.T) {
		ed := Definition[any, any]{
			table: "table",
		}
		require.Equal(t, "table", ed.RelationName())
	})

	t.Run("should produce parameterized fields", func(t *testing.T) {
		def := Definition[any, any]{
			schema: "schema",
			table:  "table",
			fields: Fields{
				{Name: "id", IsIdentity: true},
				{Name: "name"},
			},
		}

		require.Equal(t, []any{
			":id", ":name",
		}, def.Fields().FieldParams(""))
	})
}
