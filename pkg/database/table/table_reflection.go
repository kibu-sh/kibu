package table

import (
	"github.com/discernhq/devx/pkg/database/xql"
	"github.com/fatih/structtag"
	"reflect"
	"strings"
)

func MustReflect[E any](driver xql.Driver, tagName string) (def *Mapper[E]) {
	def, err := Reflect[E](driver, tagName)
	if err != nil {
		panic(err)
	}
	return
}

func Reflect[E any](driver xql.Driver, tagName string) (def *Mapper[E], err error) {
	r := reflect.TypeOf(new(E)).Elem()
	def = new(Mapper[E])
	def.Table = r.Name()
	def.StructToDB = make(map[string]string)
	def.DBToStruct = make(map[string]StructMetadata)
	def.Builder, err = xql.NewBuilder(driver)
	if err != nil {
		return
	}

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

			def.Columns = append(def.Columns, Column{
				Name:       tag.Name,
				IsIdentity: options.Has("pk"),
			})

			def.StructToDB[field.Name] = tag.Name
			def.DBToStruct[tag.Name] = StructMetadata{
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
