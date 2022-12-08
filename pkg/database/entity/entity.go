package entity

import (
	. "github.com/discernhq/devx/pkg/database/xql"
	"strings"
)

var _ Mapper = (*Definition[any, any])(nil)
var _ QueryBuilder = (*Definition[any, any])(nil)
var _ PKQueryBuilder[any, any] = (*Definition[any, any])(nil)

type Mapper interface {
	Schema() string
	Table() string
	RelationName() string
	Fields() Fields
}

type Definition[E any, PK any] struct {
	schema     string
	table      string
	fields     Fields
	structToDB map[string]string
	dbToStruct map[string]structReflectMeta
}

type structReflectMeta struct {
	Name string
	ID   int
}

func (d *Definition[E, PK]) Schema() string {
	return d.schema
}

func (d *Definition[E, PK]) Table() string {
	return d.table
}

func (d *Definition[E, PK]) Fields() Fields {
	return d.fields
}

func (d *Definition[E, PK]) SelectOneBuilder(primaryKey PK) SelectBuilder {
	return d.SelectBuilder().Where(Eq{
		d.Fields().PrimaryKey().String(): primaryKey,
	})
}

func (d *Definition[E, PK]) UpdateOneBuilder(primaryKey PK) UpdateBuilder {
	return d.UpdateBuilder().Where(Eq{
		d.Fields().PrimaryKey().String(): primaryKey,
	})
}

func (d *Definition[E, PK]) DeleteOneBuilder(primaryKey PK) DeleteBuilder {
	return d.DeleteBuilder().Where(Eq{
		d.Fields().PrimaryKey().String(): primaryKey,
	})
}

func (d *Definition[E, PK]) SelectBuilder() SelectBuilder {
	return Select(d.fields.Names()...).From(d.RelationName())
}

func (d *Definition[E, PK]) InsertBuilder() InsertBuilder {
	return Insert(d.RelationName()).Columns(d.fields.Names()...)
}

func (d *Definition[E, PK]) UpdateBuilder() UpdateBuilder {
	return Update(d.RelationName()).Table(d.RelationName())
}

func (d *Definition[E, PK]) DeleteBuilder() DeleteBuilder {
	return Delete(d.RelationName()).From(d.RelationName())
}

// RelationName returns the fully qualified name of the entity in the database.
// TODO: this may be different in other dialects. We may need to wrap these.
func (d *Definition[E, PK]) RelationName() string {
	parts := []string{d.table}
	if d.schema != "" {
		parts = append([]string{d.schema}, parts...)
	}
	return strings.Join(parts, ".")
}
