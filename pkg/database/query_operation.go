package database

type Op interface {
	Query() Query
}

type OpFindOne struct {
	query Query
}

func (o OpFindOne) Query() Query { return o.query }

type OpFindMany struct {
	query Query
}

func (o OpFindMany) Query() Query { return o.query }
