package entity

import (
	"fmt"
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

func (f Fields) FieldParams(prefix string) (names []any) {
	if prefix == "" {
		prefix = ":"
	}
	for _, name := range f.Names() {
		names = append(names, fmt.Sprintf(":%s", name))
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
