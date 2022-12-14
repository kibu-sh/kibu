package database

import (
	"context"
	"github.com/discernhq/devx/pkg/database/xql"
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
	context.Context
	Operation() Operation
	Query() xql.Query
}

type OpContext struct {
	context.Context
	operation Operation
	query     xql.Query
}

func (o OpContext) Operation() Operation {
	return o.operation
}

func (o OpContext) Query() xql.Query {
	return o.query
}
