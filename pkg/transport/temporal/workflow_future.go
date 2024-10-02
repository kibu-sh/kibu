package temporal

import "go.temporal.io/sdk/workflow"

var _ Future[any] = (*future[any])(nil)

type future[T any] struct {
	wf workflow.Future
}

func (f future[T]) Select(sel workflow.Selector, fn FutureCallback[T]) workflow.Selector {
	return sel.AddFuture(f.wf, func(workflow.Future) {
		if fn != nil {
			fn(f)
		}
	})
}

func (f future[T]) Get(ctx workflow.Context) (res T, err error) {
	err = f.wf.Get(ctx, &res)
	return
}

func (f future[T]) IsReady() bool {
	return f.wf.IsReady()
}

func (f future[T]) Underlying() workflow.Future {
	return f.wf
}

func NewFuture[T any](f workflow.Future) Future[T] {
	return future[T]{f}
}
