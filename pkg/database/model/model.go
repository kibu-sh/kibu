package model

import (
	. "github.com/discernhq/devx/pkg/database/xql"
	"reflect"
	"strings"
)

var _ Definition = (*Mapper[any])(nil)
var _ QueryBuilder = (*Mapper[any])(nil)
var _ EntityQueryBuilder[any] = (*Mapper[any])(nil)

type Definition interface {
	Schema() string
	Table() string
	RelationName() string
	Fields() Fields
}

type Mapper[E any] struct {
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

func (d *Mapper[E]) Schema() string {
	return d.schema
}

func (d *Mapper[E]) Table() string {
	return d.table
}

func (d *Mapper[E]) Fields() Fields {
	return d.fields
}

func (d *Mapper[E]) PrimaryKeyPredicate(entity *E) (result Eq) {
	result = make(map[string]any)
	reflected := reflect.ValueOf(entity).Elem()
	for _, field := range d.fields.IdentityFields() {
		meta := d.dbToStruct[field.Name]
		result[field.Name] = reflected.Field(meta.ID).Interface()
	}
	return
}

func (d *Mapper[E]) SelectOneBuilder(entity *E) SelectBuilder {
	return d.SelectBuilder().Where(d.PrimaryKeyPredicate(entity))
}

func (d *Mapper[E]) UpdateOneBuilder(entity *E) UpdateBuilder {
	return d.UpdateBuilder().Where(d.PrimaryKeyPredicate(entity))
}

func (d *Mapper[E]) DeleteOneBuilder(e *E) DeleteBuilder {
	return d.DeleteBuilder().Where(d.PrimaryKeyPredicate(e))
}

func (d *Mapper[E]) SelectBuilder() SelectBuilder {
	return Select(d.fields.Names()...).From(d.RelationName())
}

func (d *Mapper[E]) InsertBuilder() InsertBuilder {
	return Insert(d.RelationName()).Columns(d.fields.Names()...)
}

func (d *Mapper[E]) UpdateBuilder() UpdateBuilder {
	return Update(d.RelationName()).Table(d.RelationName())
}

func (d *Mapper[E]) DeleteBuilder() DeleteBuilder {
	return Delete(d.RelationName()).From(d.RelationName())
}

// RelationName returns the fully qualified name of the model in the database.
// TODO: this may be different in other dialects. We may need to wrap these.
func (d *Mapper[E]) RelationName() string {
	parts := []string{d.table}
	if d.schema != "" {
		parts = append([]string{d.schema}, parts...)
	}
	return strings.Join(parts, ".")
}

type ValueMap map[string]any

func (d *Mapper[E]) ValueMap(entity *E) (values ValueMap) {
	values = make(map[string]any)
	reflected := reflect.ValueOf(entity).Elem()
	for s, meta := range d.dbToStruct {
		values[s] = reflected.Field(meta.ID).Interface()
	}
	return
}

func (d *Mapper[E]) ValueMapToEntity(valueMap ValueMap) (entity *E) {
	entity = new(E)
	reflected := reflect.ValueOf(entity).Elem()

	for dbField, value := range valueMap {
		if meta, ok := d.dbToStruct[dbField]; ok {
			reflected.Field(meta.ID).Set(reflect.ValueOf(value))
		}
	}
	return
}

func (d *Mapper[E]) ColumnValues(entity *E) (values []any) {
	reflected := reflect.ValueOf(entity).Elem()
	// deterministic list of values by field order
	for _, field := range d.Fields() {
		meta := d.dbToStruct[field.Name]
		values = append(values, reflected.Field(meta.ID).Interface())
	}
	return
}
