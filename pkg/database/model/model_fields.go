package model

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

func (f Fields) IdentityFields() (fields Fields) {
	for _, field := range f {
		if field.IsIdentity {
			fields = append(fields, field)
		}
	}
	return
}
