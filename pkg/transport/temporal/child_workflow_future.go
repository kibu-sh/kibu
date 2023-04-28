package temporal

import "go.temporal.io/sdk/workflow"

var _ ChildWorkflowFuture[any] = (*childWorkflowFuture[any])(nil)

type childWorkflowFuture[T any] struct {
	cwf workflow.ChildWorkflowFuture
}

func (f childWorkflowFuture[T]) UnderlyingFuture() workflow.Future {
	return f.cwf
}

func (f childWorkflowFuture[T]) Get(ctx workflow.Context) (res T, err error) {
	err = f.cwf.Get(ctx, &res)
	return
}

func (f childWorkflowFuture[T]) IsReady() bool {
	return f.cwf.IsReady()
}

func (f childWorkflowFuture[T]) GetChildWorkflowExecution() Future[T] {
	return NewFuture[T](f.cwf.GetChildWorkflowExecution())
}

func (f childWorkflowFuture[T]) UnderlyingChildWorkflowFuture() workflow.ChildWorkflowFuture {
	return f.cwf
}

func NewChildWorkflowFuture[T any](f workflow.ChildWorkflowFuture) ChildWorkflowFuture[T] {
	return childWorkflowFuture[T]{f}
}
