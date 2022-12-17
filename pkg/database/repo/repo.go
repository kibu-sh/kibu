package repo

import (
	"github.com/discernhq/devx/pkg/database/model"
	"github.com/discernhq/devx/pkg/database/xql"
)

type Repo[Model any] struct {
	runner        xql.Runner
	options       *Options
	mapper        *model.Mapper[Model]
	queryPipeline HookFunc
}

type Options struct {
	Logger          Logger
	QueryHookChain  HookChain
	ExecHookChain   HookChain
	ResultHookChain HookChain
}

type HookChain []HookFunc
type OptionFunc func(options *Options) error

type Logger interface{}
type noOpLogger struct{}

func WithLogger(logger Logger) OptionFunc {
	return func(options *Options) error {
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
type ResultHookFunc func(ctx Context, result any) error
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

func WithQueryHook(hook HookFunc) OptionFunc {
	return func(options *Options) error {
		options.QueryHookChain = append(options.QueryHookChain, hook)
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

func newQueryHookChain(conn xql.Runner) HookChain {
	return HookChain{
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
	}
}

func NewRepoOrPanic[Model any](opts ...OptionFunc) (repo *Repo[Model]) {
	var err error
	repo, err = NewRepo[Model](opts...)
	if err != nil {
		panic(err)
	}
	return
}

func NewRepo[Model any](opts ...OptionFunc) (repo *Repo[Model], err error) {
	repo = &Repo[Model]{
		options: &Options{},
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

	return
}

func (r *Repo[Model]) Query(runner xql.Runner) *Query[Model] {
	return &Query[Model]{
		runner: runner,
		mapper: r.mapper,
		pipeline: applyHooks(
			noOpQueryHook,
			joinHookChains(
				r.options.ResultHookChain,
				newQueryHookChain(runner),
				r.options.QueryHookChain,
			)...,
		),
	}
}
