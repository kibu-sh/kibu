package table

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEntityDefinition(t *testing.T) {
	t.Run("should produce a composite relation name", func(t *testing.T) {
		def := Mapper[any]{
			Schema: "Schema",
			Table:  "Table",
		}
		require.Equal(t, "Schema.Table", def.RelationName())
	})

	t.Run("should exclude empty Schema from relation name", func(t *testing.T) {
		ed := Mapper[any]{
			Table: "Table",
		}
		require.Equal(t, "Table", ed.RelationName())
	})
}
