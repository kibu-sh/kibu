package table

type Column struct {
	Name       string
	IsIdentity bool
}

type Columns []Column

func (f Columns) Names() (names []string) {
	for _, field := range f {
		names = append(names, field.Name)
	}
	return
}

func (f Columns) IdentityColumns() (columns Columns) {
	for _, field := range f {
		if field.IsIdentity {
			columns = append(columns, field)
		}
	}
	return
}
