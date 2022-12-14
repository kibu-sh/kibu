package database

import (
	"context"
	"github.com/discernhq/devx/pkg/database/model"
	"github.com/discernhq/devx/pkg/database/xql"
)

var _ QueryMethods[any] = (*Repo[any])(nil)
var _ CommandMethods[any] = (*Repo[any])(nil)

type QueryMethods[Model any] interface {
	FindOne(ctx context.Context, model *Model) (result *Model, err error)
	FindMany(ctx context.Context, builder xql.SelectBuilderFunc) (result []*Model, err error)
}

type CreateMethods[Model any] interface {
	CreateOne(ctx context.Context, model *Model, builders ...xql.InsertBuilderFunc) (err error)
	CreateMany(ctx context.Context, models []*Model, builders ...xql.InsertBuilderFunc) (err error)
}

type SaveMethods[Model any] interface {
	SaveOne(ctx context.Context, model *Model, builders ...xql.UpdateBuilderFunc) (err error)
	SaveMany(ctx context.Context, models []*Model, builders ...xql.UpdateBuilderFunc) (err error)
}

type UpdateMethods[Model any] interface {
	UpdateMany(ctx context.Context, builders ...xql.UpdateBuilderFunc) (err error)
}

type DeleteMethods[Model any] interface {
	DeleteOne(ctx context.Context, model *Model, builders ...xql.DeleteBuilderFunc) (err error)
	DeleteMany(ctx context.Context, builders ...xql.DeleteBuilderFunc) (err error)
}

type CommandMethods[Model any] interface {
	SaveMethods[Model]
	CreateMethods[Model]
	UpdateMethods[Model]
	DeleteMethods[Model]
}

type Repo[Model any] struct {
	runner        xql.Runner
	options       *RepoOptions
	mapper        *model.Definition[Model]
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

func noOpQueryHook(_ Context, _ any) (err error) {
	return
}

func newFindOneHook(runner xql.Runner) HookFunc {
	return func(ctx Context, result any) error {
		return xql.QueryWith(ctx, xql.QueryWithParams{
			Target:       result,
			Query:        ctx.Query(),
			QueryStmFunc: runner.GetContext,
		})
	}
}

func newFindManyHook(runner xql.Runner) HookFunc {
	return func(ctx Context, result any) (err error) {
		return xql.QueryWith(ctx, xql.QueryWithParams{
			Target:       result,
			Query:        ctx.Query(),
			QueryStmFunc: runner.SelectContext,
		})
	}
}

func newExecHook(runner xql.Runner) HookFunc {
	return func(ctx Context, result any) (err error) {
		return xql.QueryWith(ctx, xql.QueryWithParams{
			Target:       result,
			Query:        ctx.Query(),
			QueryStmFunc: xql.ExecAsQueryStmFunc(runner.ExecContext),
		})
	}
}

func bindHookToOperations(hook HookFunc, supported map[Operation]bool) HookFunc {
	return func(ctx Context, result any) (err error) {
		if _, ok := supported[ctx.Operation()]; ok {
			err = hook(ctx, result)
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

func newDefaultRepoOptions(conn xql.Runner) *RepoOptions {
	return &RepoOptions{
		ExecHookChain: HookChain{
			bindHookToOperations(newFindOneHook(conn), map[Operation]bool{
				OpFindOne: true,
			}),
			bindHookToOperations(newFindManyHook(conn), map[Operation]bool{
				OpFindMany: true,
			}),
			bindHookToOperations(newExecHook(conn), map[Operation]bool{
				OpCreateOne:  true,
				OpSaveOne:    true,
				OpUpdateMany: true,
				OpDeleteOne:  true,
				OpDeleteMany: true,
			}),
		},
	}
}

func NewRepo[Model any](runner xql.Runner, opts ...RepoOptionFunc) (repo *Repo[Model], err error) {
	options := newDefaultRepoOptions(runner)

	repo = &Repo[Model]{
		runner:  runner,
		options: options,
	}

	repo.mapper, err = model.Reflect[Model]("db")
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

func (r *Repo[Model]) FindOne(ctx context.Context, model *Model) (result *Model, err error) {
	result = new(Model)
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpFindOne,
		query:     r.mapper.SelectOneBuilder(model),
	}, result)
	return
}

func (r *Repo[Model]) FindMany(ctx context.Context, selectBuilder xql.SelectBuilderFunc) (result []*Model, err error) {
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpFindMany,
		query:     selectBuilder(r.mapper.SelectBuilder()),
	}, &result)
	return
}

func (r *Repo[Model]) CreateOne(ctx context.Context, model *Model, builders ...xql.InsertBuilderFunc) (err error) {
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpCreateOne,
		query: xql.ApplyInsertBuilderFuncs(
			r.mapper.InsertBuilder().SetMap(r.mapper.ValueMap(model)),
			builders...,
		),
	}, model)
	return
}

func (r *Repo[Model]) CreateMany(ctx context.Context, models []*Model, builders ...xql.InsertBuilderFunc) (err error) {
	for _, m := range models {
		if err = r.CreateOne(ctx, m, builders...); err != nil {
			return
		}
	}
	return
}

func (r *Repo[Model]) SaveOne(ctx context.Context, model *Model, builders ...xql.UpdateBuilderFunc) (err error) {
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpSaveOne,
		query: xql.ApplyUpdateBuilderFuncs(
			r.mapper.UpdateOneBuilder(model).SetMap(r.mapper.ValueMap(model)),
			builders...,
		),
	}, model)
	return
}

func (r *Repo[Model]) SaveMany(ctx context.Context, models []*Model, builders ...xql.UpdateBuilderFunc) (err error) {
	for _, m := range models {
		if err = r.SaveOne(ctx, m, builders...); err != nil {
			return
		}
	}
	return
}

func (r *Repo[Model]) UpdateMany(ctx context.Context, builders ...xql.UpdateBuilderFunc) (err error) {
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpUpdateMany,
		query: xql.ApplyUpdateBuilderFuncs(
			r.mapper.UpdateBuilder(),
			builders...,
		),
	}, nil)
	return
}

func (r *Repo[Model]) DeleteOne(ctx context.Context, model *Model, builders ...xql.DeleteBuilderFunc) (err error) {
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpDeleteOne,
		query: xql.ApplyDeleteBuilderFuncs(
			r.mapper.DeleteOneBuilder(model),
			builders...,
		),
	}, nil)
	return
}

func (r *Repo[Model]) DeleteMany(ctx context.Context, builders ...xql.DeleteBuilderFunc) (err error) {
	err = r.queryPipeline(&OpContext{
		Context:   ctx,
		operation: OpDeleteMany,
		query: xql.ApplyDeleteBuilderFuncs(
			r.mapper.DeleteBuilder(),
			builders...,
		),
	}, nil)
	return
}
