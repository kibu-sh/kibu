package database

import (
	"github.com/fatih/structtag"
	"reflect"
	"strings"
)

type Field struct {
	Name       string
	IsIdentity bool
}

type Fields []Field

func (f Fields) Names() (names []string) {
	for _, field := range f {
		names = append(names, field.Name)
	}
	return
}

func (f Fields) PrimaryKey() (keys PrimaryKey) {
	for _, field := range f {
		if field.IsIdentity {
			keys = append(keys, field)
		}
	}
	return
}

type PrimaryKey []Field

func (p PrimaryKey) Names() (names []string) {
	for _, field := range p {
		names = append(names, field.Name)
	}
	return
}

func (p PrimaryKey) String() string {
	return strings.Join(p.Names(), ".")
}

func ReflectEntityDefinition[T any](tagName string) (def EntityDefinition, err error) {
	r := reflect.TypeOf(new(T)).Elem()

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
				def.Table = options.Get("table")
			}

			if options.Has("schema") {
				def.Schema = options.Get("schema")
			}

			def.Fields = append(def.Fields, Field{
				Name:       tag.Name,
				IsIdentity: options.Has("pk"),
			})
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
