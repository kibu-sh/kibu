package database

import "strings"

type EntityDefinition struct {
	Schema string
	Table  string
	Fields Fields
}

// RelationName returns the fully qualified name of the entity in the database.
// TODO: this may be different in other dialects. We may need to wrap these.
func (d EntityDefinition) RelationName() string {
	parts := []string{d.Table}
	if d.Schema != "" {
		parts = append([]string{d.Schema}, parts...)
	}
	return strings.Join(parts, ".")
}
