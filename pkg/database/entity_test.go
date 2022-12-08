package database

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEntityDefinition(t *testing.T) {
	t.Run("should produce a composite relation name", func(t *testing.T) {
		ed := EntityDefinition{
			Schema: "schema",
			Table:  "table",
		}
		require.Equal(t, "schema.table", ed.RelationName())
	})

	t.Run("should exclude empty schema from relation name", func(t *testing.T) {
		ed := EntityDefinition{
			Table: "table",
		}
		require.Equal(t, "table", ed.RelationName())
	})
}
