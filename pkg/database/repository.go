package database

import (
	"context"
	"github.com/discernhq/devx/pkg/database/entity"
	"github.com/discernhq/devx/pkg/database/xql"
)

var _ QueryMethods[any, any] = (*Repo[any, any])(nil)
var _ CommandMethods[any, any] = (*Repo[any, any])(nil)

type QueryMethods[Entity, PK any] interface {
	FindOne(ctx context.Context, primaryKey PK) (model *Entity, err error)
	FindMany(ctx context.Context, builder xql.SelectBuilderFunc) (models []*Entity, err error)
}

type CreateMethods[Entity, PK any] interface {
	CreateOne(ctx context.Context, entity *Entity) (err error)
	CreateMany(ctx context.Context, entities []*Entity) (err error)
}

type SaveMethods[Entity, PK any] interface {
	SaveOne(ctx context.Context, model *Entity) (err error)
	SaveMany(ctx context.Context, models []*Entity) (err error)
}

type UpdateMethods[Entity, PK any] interface {
	UpdateOne(ctx context.Context, pk PK, builder xql.UpdateBuilderFunc) (err error)
	UpdateMany(ctx context.Context, builder xql.UpdateBuilderFunc) (err error)
}

type DeleteMethods[Entity, PK any] interface {
	DeleteOne(ctx context.Context, pk PK) (err error)
	DeleteMany(ctx context.Context, builder xql.DeleteBuilderFunc) (err error)
}

type CommandMethods[Entity, PK any] interface {
	SaveMethods[Entity, PK]
	CreateMethods[Entity, PK]
	UpdateMethods[Entity, PK]
	DeleteMethods[Entity, PK]
}

type Repo[Entity, PK any] struct {
	conn          Conn
	options       *RepoOptions
	mapper        *entity.Definition[Entity, PK]
	queryPipeline HookFunc
}

type RepoOptions struct {
	Logger           Logger
	QueryHookChain   HookChain
	ExecHookChain    HookChain
	PrivacyHookChain HookChain
}

type HookChain []HookFunc
type RepoOptionFunc func(options *RepoOptions) error

type Logger interface{}
type noOpLogger struct{}

func WithLogger(logger Logger) RepoOptionFunc {
	return func(options *RepoOptions) error {
		options.Logger = logger
		return nil
	}
}

func noOpQueryHook(ctx Context, result any) (err error) { return }

func newFindOneHook(conn Conn) HookFunc {
	return func(ctx Context, result any) (err error) {
		if ctx.Operation() == OpFindOne {
			err = conn.Get(ctx, result, ctx.Query())
		}
		return
	}
}

func newFindManyHook(conn Conn) HookFunc {
	return func(ctx Context, result any) (err error) {
		if ctx.Operation() == OpFindMany {
			err = conn.Select(ctx, result, ctx.Query())
		}
		return
	}
}

func newExecHook(conn Conn) HookFunc {
	supported := map[Operation]bool{
		OpSaveOne:    true,
		OpCreateOne:  true,
		OpCreateMany: true,
		OpUpdateOne:  true,
		OpUpdateMany: true,
		OpDeleteOne:  true,
		OpDeleteMany: true,
	}
	return func(ctx Context, result any) (err error) {
		if _, ok := supported[ctx.Operation()]; ok {
			_, err = conn.Exec(ctx, ctx.Query())
		}
		return
	}
}

func applyHooks(base HookFunc, chain ...HookFunc) (result HookFunc) {
	result = base
	for _, hook := range chain {
		result = HookDecorator(hook)(result)
	}
	return
}

type HookFunc func(ctx Context, result any) error

type PrivacyHookFunc func(ctx Context, result any) error
type HookDecoratorFunc func(next HookFunc) HookFunc

func HookDecorator(base HookFunc) HookDecoratorFunc {
	return func(next HookFunc) HookFunc {
		return func(ctx Context, result any) error {
			if err := base(ctx, result); err != nil {
				return err
			}
			return next(ctx, result)
		}
	}
}

func WithQueryHook(hook HookFunc) RepoOptionFunc {
	return func(options *RepoOptions) error {
		options.QueryHookChain = append(options.QueryHookChain, hook)
		return nil
	}
}

func WithPrivacyHook(privacyHook PrivacyHookFunc) RepoOptionFunc {
	return func(options *RepoOptions) error {
		options.PrivacyHookChain = append(
			options.PrivacyHookChain,
			HookFunc(privacyHook),
		)
		return nil
	}
}

func joinHookChains(chains ...HookChain) (result HookChain) {
	for _, chain := range chains {
		result = append(result, chain...)
	}
	return
}

func NewRepo[Entity, PK any](conn Conn, opts ...RepoOptionFunc) (repo *Repo[Entity, PK], err error) {
	options := &RepoOptions{
		ExecHookChain: HookChain{
			newFindOneHook(conn),
			newFindManyHook(conn),
			newExecHook(conn),
		},
	}

	repo = &Repo[Entity, PK]{
		conn:    conn,
		options: options,
	}

	repo.mapper, err = entity.ReflectEntityDefinition[Entity, PK]("db")
	if err != nil {
		return
	}

	for _, opt := range opts {
		if err = opt(repo.options); err != nil {
			return
		}
	}

	repo.queryPipeline = applyHooks(
		noOpQueryHook,
		joinHookChains(
			repo.options.PrivacyHookChain,
			repo.options.ExecHookChain,
			repo.options.QueryHookChain,
		)...,
	)

	return
}

func (r *Repo[Entity, PK]) FindOne(ctx context.Context, primaryKey PK) (entity *Entity, err error) {
	entity = new(Entity)
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpFindOne,
		query:     r.mapper.SelectOneBuilder(primaryKey),
	}, entity)
	return
}

func (r *Repo[Entity, PK]) FindMany(ctx context.Context, selectBuilder xql.SelectBuilderFunc) (entities []*Entity, err error) {
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpFindMany,
		query:     selectBuilder(r.mapper.SelectBuilder()),
	}, &entities)
	return
}

func (r *Repo[Entity, PK]) CreateOne(ctx context.Context, entity *Entity) (err error) {
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpCreateOne,
		query:     r.mapper.InsertBuilder().SetMap(r.mapper.ValueMap(entity)),
	}, entity)
	return
}

func (r *Repo[Entity, PK]) CreateMany(ctx context.Context, entity []*Entity) (err error) {
	query := r.mapper.InsertBuilder()
	for _, e := range entity {
		query = query.Values(r.mapper.ColumnValues(e)...)
	}
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpCreateMany,
		query:     query,
	}, entity)
	return
}

func (r *Repo[Entity, PK]) SaveOne(ctx context.Context, entity *Entity) (err error) {
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpSaveOne,
		query: r.mapper.UpdateOneBuilder(
			r.mapper.PrimaryKey(entity),
		).SetMap(r.mapper.ValueMap(entity)),
	}, entity)
	return
}

func (r *Repo[Entity, PK]) SaveMany(ctx context.Context, entities []*Entity) (err error) {
	for _, entity := range entities {
		if err = r.SaveOne(ctx, entity); err != nil {
			return
		}
	}
	return
}

func (r *Repo[Entity, PK]) UpdateOne(ctx context.Context, primaryKey PK, updateBuilder xql.UpdateBuilderFunc) (err error) {
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpUpdateOne,
		query:     updateBuilder(r.mapper.UpdateOneBuilder(primaryKey)),
	}, nil)
	return
}

func (r *Repo[Entity, PK]) UpdateMany(ctx context.Context, updateBuilder xql.UpdateBuilderFunc) (err error) {
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpUpdateMany,
		query:     updateBuilder(r.mapper.UpdateBuilder()),
	}, nil)
	return
}

func (r *Repo[Entity, PK]) DeleteOne(ctx context.Context, pk PK) (err error) {
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpDeleteOne,
		query:     r.mapper.DeleteOneBuilder(pk),
	}, nil)
	return
}

func (r *Repo[Entity, PK]) DeleteMany(ctx context.Context, deleteBuilder xql.DeleteBuilderFunc) (err error) {
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpDeleteMany,
		query:     deleteBuilder(r.mapper.DeleteBuilder()),
	}, nil)
	return
}

// func (r *Repo[Entity, PK]) Count(ctx context.Context, selectBuilder xql.SelectBuilderFunc) (count int, err error) {
// 	err = r.queryPipeline(&OpContext{
// 		Context:   ctx,
// 		operation: OpCount,
// 		query:     selectBuilder(r.mapper.CountBuilder()),
// 	}, &count)
// 	return
// }
