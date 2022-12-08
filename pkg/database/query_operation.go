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
	OpCreateMany
	OpSaveOne
	OpSaveMany
	OpUpdateOne
	OpUpdateMany
	OpDeleteOne
	OpDeleteMany
)

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
