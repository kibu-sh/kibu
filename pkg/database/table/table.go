package table

import (
	"github.com/discernhq/devx/pkg/database/xql"
	"reflect"
	"strings"
)

var _ Definition = (*Mapper[any])(nil)
var _ xql.QueryBuilder = (*Mapper[any])(nil)
var _ xql.EntityQueryBuilder[any] = (*Mapper[any])(nil)

type Definition interface {
	SchemaName() string
	TableName() string
	RelationName() string
	ColumnList() Columns
}

type Mapper[E any] struct {
	Schema     string
	Table      string
	Columns    Columns
	Builder    xql.StatementBuilderType
	StructToDB map[string]string
	DBToStruct map[string]StructMetadata
}

type StructMetadata struct {
	Name string
	ID   int
}

func (d Mapper[E]) WithTableName(name string) *Mapper[E] {
	d.Table = name
	return &d
}

func (d Mapper[E]) WithSchemaName(name string) *Mapper[E] {
	d.Schema = name
	return &d
}

func (d Mapper[E]) SchemaName() string {
	return d.Schema
}

func (d Mapper[E]) TableName() string {
	return d.Table
}

func (d Mapper[E]) ColumnList() Columns {
	return d.Columns
}

func (d Mapper[E]) PrimaryKeyPredicate(entity *E) (result xql.Eq) {
	result = make(map[string]any)
	reflected := reflect.ValueOf(entity).Elem()
	for _, column := range d.Columns.IdentityColumns() {
		meta := d.DBToStruct[column.Name]
		result[column.Name] = reflected.Field(meta.ID).Interface()
	}
	return
}

func (d Mapper[E]) SelectOneBuilder(entity *E) xql.SelectBuilder {
	return d.SelectBuilder().Where(d.PrimaryKeyPredicate(entity))
}

func (d Mapper[E]) UpdateOneBuilder(entity *E) xql.UpdateBuilder {
	return d.UpdateBuilder().Where(d.PrimaryKeyPredicate(entity))
}

func (d Mapper[E]) DeleteOneBuilder(e *E) xql.DeleteBuilder {
	return d.DeleteBuilder().Where(d.PrimaryKeyPredicate(e))
}

func (d Mapper[E]) SelectBuilder() xql.SelectBuilder {
	return d.Builder.Select(d.Columns.Names()...).From(d.RelationName())
}

func (d Mapper[E]) InsertBuilder() xql.InsertBuilder {
	return d.Builder.Insert(d.RelationName()).Columns(d.Columns.Names()...)
}

func (d Mapper[E]) UpdateBuilder() xql.UpdateBuilder {
	return d.Builder.Update(d.RelationName()).Table(d.RelationName())
}

func (d Mapper[E]) DeleteBuilder() xql.DeleteBuilder {
	return d.Builder.Delete(d.RelationName()).From(d.RelationName())
}

// RelationName returns the fully qualified name of the Table in the database.
// TODO: this may be different in other dialects. We may need to wrap these.
func (d Mapper[E]) RelationName() string {
	parts := []string{d.Table}
	if d.Schema != "" {
		parts = append([]string{d.Schema}, parts...)
	}
	return strings.Join(parts, ".")
}

type ValueMap map[string]any

func (d Mapper[E]) ValueMap(entity *E) (values ValueMap) {
	values = make(map[string]any)
	reflected := reflect.ValueOf(entity).Elem()
	for s, meta := range d.DBToStruct {
		values[s] = reflected.Field(meta.ID).Interface()
	}
	return
}

func (d Mapper[E]) ValueMapToModel(valueMap ValueMap) (entity *E) {
	entity = new(E)
	reflected := reflect.ValueOf(entity).Elem()

	for dbField, value := range valueMap {
		if meta, ok := d.DBToStruct[dbField]; ok {
			reflected.Field(meta.ID).Set(reflect.ValueOf(value))
		}
	}
	return
}

func (d Mapper[E]) ColumnValues(entity *E) (values []any) {
	reflected := reflect.ValueOf(entity).Elem()
	// deterministic list of values by field order
	for _, field := range d.ColumnList() {
		meta := d.DBToStruct[field.Name]
		values = append(values, reflected.Field(meta.ID).Interface())
	}
	return
}
