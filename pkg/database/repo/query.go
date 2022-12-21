package repo

import (
	"context"
	"github.com/discernhq/devx/pkg/database/table"
	"github.com/discernhq/devx/pkg/database/xql"
)

var _ QueryMethods[any] = (*Query[any])(nil)
var _ CommandMethods[any] = (*Query[any])(nil)

type QueryMethods[Model any] interface {
	FindOne(ctx context.Context, model *Model) (result *Model, err error)
	FindMany(ctx context.Context, queryFunc xql.SelectBuilderFunc) (result []*Model, err error)
}

type CreateMethods[Model any] interface {
	CreateOne(ctx context.Context, model *Model) (err error)
	CreateMany(ctx context.Context, models []*Model) (err error)
}

type SaveMethods[Model any] interface {
	SaveOne(ctx context.Context, model *Model) (err error)
	SaveMany(ctx context.Context, models []*Model) (err error)
}

type UpdateMethods[Model any] interface {
	UpdateMany(ctx context.Context, queryFunc xql.UpdateBuilderFunc) (err error)
}

type DeleteMethods[Model any] interface {
	DeleteOne(ctx context.Context, model *Model) (err error)
	DeleteMany(ctx context.Context, queryFund xql.DeleteBuilderFunc) (err error)
}

type CommandMethods[Model any] interface {
	SaveMethods[Model]
	CreateMethods[Model]
	UpdateMethods[Model]
	DeleteMethods[Model]
}

type Query[Model any] struct {
	connection xql.Connection
	pipeline   HookFunc
	mapper     *table.Mapper[Model]
}

type QueryProviderFunc[Model any] func(ctx context.Context) (*Query[Model], error)

func (q *Query[Model]) FindOne(ctx context.Context, model *Model) (result *Model, err error) {
	result = new(Model)
	err = q.pipeline(&OpContext{
		Context:   ctx,
		operation: OpFindOne,
		query:     q.mapper.SelectOneBuilder(model),
	}, result)
	return
}

func (q *Query[Model]) FindMany(ctx context.Context, queryFunc xql.SelectBuilderFunc) (result []*Model, err error) {
	err = q.pipeline(&OpContext{
		Context:   ctx,
		operation: OpFindMany,
		query:     queryFunc(q.mapper.SelectBuilder()),
	}, &result)
	return
}

func (q *Query[Model]) CreateOne(ctx context.Context, model *Model) (err error) {
	err = q.pipeline(&OpContext{
		Context:   ctx,
		operation: OpCreateOne,
		query:     q.mapper.InsertBuilder().SetMap(q.mapper.ValueMap(model)),
	}, model)
	return
}

func (q *Query[Model]) CreateMany(ctx context.Context, models []*Model) (err error) {
	for _, m := range models {
		if err = q.CreateOne(ctx, m); err != nil {
			return
		}
	}
	return
}

func (q *Query[Model]) SaveOne(ctx context.Context, model *Model) (err error) {
	err = q.pipeline(&OpContext{
		Context:   ctx,
		operation: OpSaveOne,
		query:     q.mapper.UpdateOneBuilder(model).SetMap(q.mapper.ValueMap(model)),
	}, model)
	return
}

func (q *Query[Model]) SaveMany(ctx context.Context, models []*Model) (err error) {
	for _, m := range models {
		if err = q.SaveOne(ctx, m); err != nil {
			return
		}
	}
	return
}

func (q *Query[Model]) UpdateMany(ctx context.Context, queryFunc xql.UpdateBuilderFunc) (err error) {
	err = q.pipeline(&OpContext{
		Context:   ctx,
		operation: OpUpdateMany,
		query:     queryFunc(q.mapper.UpdateBuilder()),
	}, nil)
	return
}

func (q *Query[Model]) DeleteOne(ctx context.Context, model *Model) (err error) {
	err = q.pipeline(&OpContext{
		Context:   ctx,
		operation: OpDeleteOne,
		query:     q.mapper.DeleteOneBuilder(model),
	}, nil)
	return
}

func (q *Query[Model]) DeleteMany(ctx context.Context, queryFunc xql.DeleteBuilderFunc) (err error) {
	err = q.pipeline(&OpContext{
		Context:   ctx,
		operation: OpDeleteMany,
		query:     queryFunc(q.mapper.DeleteBuilder()),
	}, nil)
	return
}
