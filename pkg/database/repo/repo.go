package repo

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/database/table"
	"github.com/kibu-sh/kibu/pkg/database/xql"
)

type Repo[Model any] struct {
	Options *Options
	Mapper  *table.Mapper[Model]
}

type Options struct {
	Logger          Logger
	QueryHookChain  HookChain
	ResultHookChain HookChain
}

type Logger interface{}
type NoOpLogger struct{}
type OptionFunc func(options *Options) error

type HookChain []HookFunc
type HookFunc func(ctx Context, result any) error
type ResultHookFunc func(ctx Context, result any) error
type HookDecoratorFunc func(next HookFunc) HookFunc

func WithLogger(logger Logger) OptionFunc {
	return func(options *Options) error {
		options.Logger = logger
		return nil
	}
}

func NoOpQueryHook(_ Context, _ any) (err error) {
	return
}

func newFindOneHook(runner xql.Connection) HookFunc {
	return func(ctx Context, result any) error {
		return xql.QueryWith(ctx, xql.QueryWithParams{
			Target:       result,
			Query:        ctx.Query(),
			QueryStmFunc: runner.GetContext,
		})
	}
}

func newFindManyHook(runner xql.Connection) HookFunc {
	return func(ctx Context, result any) (err error) {
		return xql.QueryWith(ctx, xql.QueryWithParams{
			Target:       result,
			Query:        ctx.Query(),
			QueryStmFunc: runner.SelectContext,
		})
	}
}

func newExecHook(runner xql.Connection) HookFunc {
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

func WithQueryHook(hook HookFunc) OptionFunc {
	return func(options *Options) error {
		options.QueryHookChain = append(
			options.QueryHookChain,
			hook,
		)
		return nil
	}
}

func WithResultHook(resultHook ResultHookFunc) OptionFunc {
	return func(options *Options) error {
		options.ResultHookChain = append(
			options.ResultHookChain,
			HookFunc(resultHook),
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

func newQueryHookChain(connection xql.Connection) HookChain {
	return HookChain{
		bindHookToOperations(
			newFindOneHook(connection),
			map[Operation]bool{
				OpFindOne: true,
			},
		),
		bindHookToOperations(
			newFindManyHook(connection),
			map[Operation]bool{
				OpFindMany: true,
			},
		),
		bindHookToOperations(
			newExecHook(connection),
			map[Operation]bool{
				OpCreateOne:  true,
				OpSaveOne:    true,
				OpUpdateMany: true,
				OpDeleteOne:  true,
				OpDeleteMany: true,
			},
		),
	}
}

func NewRepoOrPanic[Model any](opts ...OptionFunc) (repo *Repo[Model]) {
	var err error
	repo, err = NewRepo[Model](xql.SQLite3, opts...)
	if err != nil {
		panic(err)
	}
	return
}

func NewRepo[Model any](driver xql.Driver, opts ...OptionFunc) (repo *Repo[Model], err error) {
	repo = &Repo[Model]{
		Options: &Options{},
	}

	repo.Mapper, err = table.Reflect[Model](driver, "db")
	if err != nil {
		return
	}

	for _, opt := range opts {
		if err = opt(repo.Options); err != nil {
			return
		}
	}

	return
}

func (r *Repo[Model]) Query(connection xql.Connection) *Query[Model] {
	return &Query[Model]{
		connection: connection,
		mapper:     r.Mapper,
		pipeline: applyHooks(
			NoOpQueryHook,
			joinHookChains(
				r.Options.ResultHookChain,
				newQueryHookChain(connection),
				r.Options.QueryHookChain,
			)...,
		),
	}
}

func (r *Repo[Model]) QueryFromCtx(ctx context.Context) (*Query[Model], error) {
	connection, err := xql.ConnectionContextStore.Load(ctx)
	if err != nil {
		return nil, err
	}
	return r.Query(connection), nil
}
