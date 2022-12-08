package entity

import (
	"github.com/fatih/structtag"
	"reflect"
	"strings"
)

func ReflectEntityDefinition[E any, PK any](tagName string) (def *Definition[E, PK], err error) {
	r := reflect.TypeOf(new(E)).Elem()
	def = new(Definition[E, PK])
	def.structToDB = make(map[string]string)
	def.dbToStruct = make(map[string]structReflectMeta)

	for i := 0; i < r.NumField(); i++ {
		field := r.Field(i)
		tags, err := structtag.Parse(string(field.Tag))
		if err != nil {
			return def, err
		}

		tag, _ := tags.Get(tagName)
		if tag == nil {
			tag = &structtag.Tag{
				Key:  tagName,
				Name: field.Name,
			}
		}
		if tag.Name != "-" {
			options := parseTagOptions(tag)

			if options.Has("table") {
				def.table = options.Get("table")
			}

			if options.Has("schema") {
				def.schema = options.Get("schema")
			}

			def.fields = append(def.fields, Field{
				Name:       tag.Name,
				IsIdentity: options.Has("pk"),
			})

			def.structToDB[field.Name] = tag.Name
			def.dbToStruct[tag.Name] = structReflectMeta{
				Name: field.Name,
				ID:   i,
			}
		}
	}
	return
}

type tagOption struct {
	key   string
	value string
}

type tagOptions []tagOption

func (t tagOptions) Has(key string) bool {
	for _, opt := range t {
		if opt.key == key {
			return true
		}
	}
	return false
}

func (t tagOptions) Get(key string) string {
	for _, opt := range t {
		if opt.key == key {
			return opt.value
		}
	}
	return ""
}

func parseTagOptions(tag *structtag.Tag) (options tagOptions) {
	for _, option := range tag.Options {
		parts := strings.Split(option, "=")
		parsed := tagOption{key: parts[0]}
		if len(parts) > 1 {
			parsed.value = parts[1]
		}
		options = append(options, parsed)
	}
	return
}

type ValueMap map[string]any

func (d *Definition[E, PK]) ValueMap(entity *E) (values ValueMap) {
	values = make(map[string]any)
	reflected := reflect.ValueOf(entity).Elem()
	for s, meta := range d.dbToStruct {
		values[s] = reflected.Field(meta.ID).Interface()
	}
	return
}

func (d *Definition[E, PK]) ValueMapToEntity(valueMap ValueMap) (entity *E) {
	entity = new(E)
	reflected := reflect.ValueOf(entity).Elem()

	for dbField, value := range valueMap {
		if meta, ok := d.dbToStruct[dbField]; ok {
			reflected.Field(meta.ID).Set(reflect.ValueOf(value))
		}
	}
	return
}

func (d *Definition[E, PK]) ColumnValues(entity *E) (values []any) {
	reflected := reflect.ValueOf(entity).Elem()
	// deterministic list of values by field order
	for _, field := range d.Fields() {
		meta := d.dbToStruct[field.Name]
		values = append(values, reflected.Field(meta.ID).Interface())
	}
	return
}