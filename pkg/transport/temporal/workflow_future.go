package temporal

import "go.temporal.io/sdk/workflow"

var _ Future[any] = (*workflowFuture[any])(nil)

type workflowFuture[T any] struct {
	wf workflow.Future
}

func (f workflowFuture[T]) Get(ctx workflow.Context) (res T, err error) {
	err = f.wf.Get(ctx, &res)
	return
}

func (f workflowFuture[T]) IsReady() bool {
	return f.wf.IsReady()
}

func (f workflowFuture[T]) UnderlyingFuture() workflow.Future {
	return f.wf
}

func NewFuture[T any](f workflow.Future) Future[T] {
	return workflowFuture[T]{f}
}
