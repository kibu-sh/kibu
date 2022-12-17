package model

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEntityDefinition(t *testing.T) {
	t.Run("should produce a composite relation name", func(t *testing.T) {
		def := Mapper[any]{
			schema: "schema",
			table:  "table",
		}
		require.Equal(t, "schema.table", def.RelationName())
	})

	t.Run("should exclude empty schema from relation name", func(t *testing.T) {
		ed := Mapper[any]{
			table: "table",
		}
		require.Equal(t, "table", ed.RelationName())
	})
}
