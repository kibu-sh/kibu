package templates

import (
	"bytes"
	"github.com/huandu/xstrings"
	"text/template"
)

var templateStdLib = template.FuncMap{
	"to_snake": xstrings.ToSnakeCase,
	"to_camel": xstrings.ToCamelCase,
	"to_kebab": xstrings.ToKebabCase,
}

type ExecFunc[T any] func(*T) (rendered bytes.Buffer, err error)

type Options struct {
	Name     string
	Contents string
	FuncMap  template.FuncMap
}

func NewTemplateExecutor[T any](tmpl *template.Template) ExecFunc[T] {
	return func(v *T) (rendered bytes.Buffer, err error) {
		err = tmpl.Execute(&rendered, v)
		return
	}
}

func DefaultOptions(name, contents string) Options {
	return Options{
		Name:     name,
		Contents: contents,
		FuncMap:  templateStdLib,
	}
}

func MustParse[T any](opts Options) ExecFunc[T] {
	return NewTemplateExecutor[T](
		template.Must(template.New(opts.Name).
			Funcs(templateStdLib).Parse(opts.Contents)),
	)
}
