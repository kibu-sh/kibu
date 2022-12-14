package model

import (
	. "github.com/discernhq/devx/pkg/database/xql"
	"reflect"
	"strings"
)

var _ Mapper = (*Definition[any])(nil)
var _ QueryBuilder = (*Definition[any])(nil)
var _ EntityQueryBuilder[any] = (*Definition[any])(nil)

type Mapper interface {
	Schema() string
	Table() string
	RelationName() string
	Fields() Fields
}

type Definition[E any] struct {
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

func (d *Definition[E]) Schema() string {
	return d.schema
}

func (d *Definition[E]) Table() string {
	return d.table
}

func (d *Definition[E]) Fields() Fields {
	return d.fields
}

func (d *Definition[E]) PrimaryKeyPredicate(entity *E) (result Eq) {
	result = make(map[string]any)
	reflected := reflect.ValueOf(entity).Elem()
	for _, field := range d.fields.IdentityFields() {
		meta := d.dbToStruct[field.Name]
		result[field.Name] = reflected.Field(meta.ID).Interface()
	}
	return
}

func (d *Definition[E]) SelectOneBuilder(entity *E) SelectBuilder {
	return d.SelectBuilder().Where(d.PrimaryKeyPredicate(entity))
}

func (d *Definition[E]) UpdateOneBuilder(entity *E) UpdateBuilder {
	return d.UpdateBuilder().Where(d.PrimaryKeyPredicate(entity))
}

func (d *Definition[E]) DeleteOneBuilder(e *E) DeleteBuilder {
	return d.DeleteBuilder().Where(d.PrimaryKeyPredicate(e))
}

func (d *Definition[E]) SelectBuilder() SelectBuilder {
	return Select(d.fields.Names()...).From(d.RelationName())
}

func (d *Definition[E]) InsertBuilder() InsertBuilder {
	return Insert(d.RelationName()).Columns(d.fields.Names()...)
}

func (d *Definition[E]) UpdateBuilder() UpdateBuilder {
	return Update(d.RelationName()).Table(d.RelationName())
}

func (d *Definition[E]) DeleteBuilder() DeleteBuilder {
	return Delete(d.RelationName()).From(d.RelationName())
}

// RelationName returns the fully qualified name of the model in the database.
// TODO: this may be different in other dialects. We may need to wrap these.
func (d *Definition[E]) RelationName() string {
	parts := []string{d.table}
	if d.schema != "" {
		parts = append([]string{d.schema}, parts...)
	}
	return strings.Join(parts, ".")
}

type ValueMap map[string]any

func (d *Definition[E]) ValueMap(entity *E) (values ValueMap) {
	values = make(map[string]any)
	reflected := reflect.ValueOf(entity).Elem()
	for s, meta := range d.dbToStruct {
		values[s] = reflected.Field(meta.ID).Interface()
	}
	return
}

func (d *Definition[E]) ValueMapToEntity(valueMap ValueMap) (entity *E) {
	entity = new(E)
	reflected := reflect.ValueOf(entity).Elem()

	for dbField, value := range valueMap {
		if meta, ok := d.dbToStruct[dbField]; ok {
			reflected.Field(meta.ID).Set(reflect.ValueOf(value))
		}
	}
	return
}

func (d *Definition[E]) ColumnValues(entity *E) (values []any) {
	reflected := reflect.ValueOf(entity).Elem()
	// deterministic list of values by field order
	for _, field := range d.Fields() {
		meta := d.dbToStruct[field.Name]
		values = append(values, reflected.Field(meta.ID).Interface())
	}
	return
}
