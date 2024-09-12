package repo

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/database/xql"
)

type Operation int

const (
	OpFindOne Operation = iota
	OpFindMany
	OpCreateOne
	OpSaveOne
	OpUpdateMany
	OpDeleteOne
	OpDeleteMany
	OpEnd
)

var operationNames = map[Operation]string{
	OpFindOne:    "FIND_ONE",
	OpFindMany:   "FIND_MANY",
	OpCreateOne:  "CREATE_ONE",
	OpSaveOne:    "SAVE_ONE",
	OpUpdateMany: "UPDATE_MANY",
	OpDeleteOne:  "DELETE_ONE",
	OpDeleteMany: "DELETE_MANY",
}

func (o Operation) String() string {
	name, ok := operationNames[o]
	if !ok {
		return "UNKNOWN"
	}
	return name
}

type Context interface {
	// FIXME: this is a ticking time bomb
	// embeding context causes stack overflow
	context.Context
	Operation() Operation
	Query() xql.StatementBuilder
	SetQuery(query xql.StatementBuilder)
}

type OpContext struct {
	context.Context
	operation Operation
	query     xql.StatementBuilder
}

func (o *OpContext) Operation() Operation {
	return o.operation
}

func (o *OpContext) Query() xql.StatementBuilder {
	return o.query
}

func (o *OpContext) SetQuery(query xql.StatementBuilder) {
	o.query = query
	return
}
