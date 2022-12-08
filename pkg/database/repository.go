package database

import (
	"context"
	"github.com/discernhq/devx/pkg/database/xql"
	"github.com/pkg/errors"
)

type QueryMethods[Entity, PK any] interface {
	FindOne(ctx context.Context, primaryKey string) (model *Entity, err error)
	FindMany(ctx context.Context, builder SelectBuilderFunc) (models []*Entity, err error)
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
}

type RepoOptions struct {
	Logger           Logger
	QueryHook        HookFunc
	EntityDefinition EntityDefinition
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

func noOpQueryHook(ctx context.Context, op Op, result any) error { return nil }

func withExecHook(conn Conn) HookFunc {
	return func(ctx context.Context, op Op, result any) (err error) {
		switch op.(type) {
		case OpFindOne:
			err = conn.Get(ctx, result, op.Query())
			break
		case OpFindMany:
			err = conn.Select(ctx, result, op.Query())
			break
		default:
			err = errors.Errorf("unsupported operation %T", op)
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

	options.EntityDefinition, err = ReflectEntityDefinition[Entity]("db")
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

type Query interface {
	ToSql() (stm string, args []any, err error)
}

type SelectBuilderFunc func(q xql.SelectBuilder) Query
type UpdateBuilderFunc func(q xql.UpdateBuilder) Query
type DeleteBuilderFunc func(q xql.DeleteBuilder) Query

type HookFunc func(ctx context.Context, op Op, result any) error
type HookDecoratorFunc func(next HookFunc) HookFunc

func HookDecorator(base HookFunc) HookDecoratorFunc {
	return func(next HookFunc) HookFunc {
		return func(ctx context.Context, op Op, result any) error {
			if err := base(ctx, op, result); err != nil {
				return err
			}
			return next(ctx, op, result)
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
	query := xql.
		Select(r.options.EntityDefinition.Fields.Names()...).
		From(r.options.EntityDefinition.RelationName()).
		Where(xql.Eq{r.options.EntityDefinition.Fields.PrimaryKey().String(): primaryKey})

	err = r.options.QueryHook(ctx, OpFindOne{
		query: query,
	}, entity)
	return
}

func (r *Repo[Entity, PK]) FindMany(ctx context.Context, selectBuilder SelectBuilderFunc) (entities []*Entity, err error) {
	query := selectBuilder(xql.
		Select(r.options.EntityDefinition.Fields.Names()...).
		From(r.options.EntityDefinition.RelationName()))

	err = r.options.QueryHook(ctx, OpFindMany{
		query: query,
	}, &entities)
	return
}
