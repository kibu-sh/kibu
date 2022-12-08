package database

import (
	"context"
	"github.com/discernhq/devx/pkg/database/entity"
	"github.com/discernhq/devx/pkg/database/xql"
	"github.com/pkg/errors"
)

type QueryMethods[Entity, PK any] interface {
	FindOne(ctx context.Context, primaryKey string) (model *Entity, err error)
	FindMany(ctx context.Context, builder xql.SelectBuilderFunc) (models []*Entity, err error)
}

type CommandMethods[Entity, PK any] interface {
	// Save(ctx context.Context, model *Entity) (err error)
	// SaveMany(ctx context.Context, models []*Entity) (err error)
	// UpdateOne(ctx context.Context, pk PK, builder UpdateBuilderFunc) (err error)
	// UpdateMany(ctx context.Context, builder UpdateBuilderFunc) (err error)

	// DeleteOne(ctx context.Context, pk PK) (err error)
	// DeleteMany(ctx context.Context, builder DeleteBuilderFunc) (err error)
}

type Repo[Entity, PK any] struct {
	conn    Conn
	options *RepoOptions
	mapper  *entity.Definition[Entity, PK]
}

type RepoOptions struct {
	Logger    Logger
	QueryHook HookFunc
}

type RepoOptionFunc func(options *RepoOptions) error

type Logger interface{}
type noOpLogger struct{}

func WithLogger(logger Logger) RepoOptionFunc {
	return func(options *RepoOptions) error {
		options.Logger = logger
		return nil
	}
}

func noOpQueryHook(ctx context.Context, op Context, result any) error { return nil }

func withExecHook(conn Conn) HookFunc {
	return func(ctx Context, result any) (err error) {
		switch ctx.Operation() {
		case OpFindOne:
			err = conn.Get(ctx, result, ctx.Query())
			break
		case OpFindMany:
			err = conn.Select(ctx, result, ctx.Query())
			break
		default:
			err = errors.Errorf("unsupported operation %T", ctx.Operation())
		}
		return
	}
}

func NewRepo[Entity, PK any](conn Conn, opts ...RepoOptionFunc) (repo *Repo[Entity, PK], err error) {
	options := &RepoOptions{
		QueryHook: withExecHook(conn),
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
	return
}

type HookFunc func(ctx Context, result any) error
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

func WithHook(hook HookFunc) RepoOptionFunc {
	return func(options *RepoOptions) error {
		options.QueryHook = HookDecorator(hook)(options.QueryHook)
		return nil
	}
}

func (r *Repo[Entity, PK]) FindOne(ctx context.Context, primaryKey PK) (entity *Entity, err error) {
	entity = new(Entity)
	err = r.options.QueryHook(&OpContext{
		Context:   ctx,
		operation: OpFindOne,
		query:     r.mapper.SelectOneBuilder(primaryKey),
	}, entity)
	return
}

func (r *Repo[Entity, PK]) FindMany(ctx context.Context, selectBuilder xql.SelectBuilderFunc) (entities []*Entity, err error) {
	err = r.options.QueryHook(&OpContext{
		Context:   ctx,
		operation: OpFindMany,
		query:     r.mapper.SelectBuilder(),
	}, &entities)
	return
}

func (r *Repo[Entity, PK]) CreateOne(ctx context.Context, entity *Entity) (err error) {
	err = r.options.QueryHook(&OpContext{
		Context:   ctx,
		operation: OpCreateOne,
		query:     r.mapper.InsertBuilder().SetMap(r.mapper.ValueMap(entity)),
	}, entity)
	return
}
