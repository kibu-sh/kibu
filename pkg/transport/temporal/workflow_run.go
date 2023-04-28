package temporal

import (
	"context"
	"go.temporal.io/sdk/client"
)

var _ WorkflowRun[any] = (*workflowRun[any])(nil)

type workflowRun[T any] struct {
	wfr client.WorkflowRun
}

func (w workflowRun[T]) GetID() string {
	return w.wfr.GetID()
}

func (w workflowRun[T]) GetRunID() string {
	return w.wfr.GetRunID()
}

func (w workflowRun[T]) Get(ctx context.Context) (result T, err error) {
	err = w.wfr.Get(ctx, &result)
	return
}

func (w workflowRun[T]) GetWithOptions(ctx context.Context, options client.WorkflowRunGetOptions) (result T, err error) {
	err = w.wfr.GetWithOptions(ctx, &result, options)
	return
}

func NewWorkflowRun[T any](run client.WorkflowRun) WorkflowRun[T] {
	return workflowRun[T]{run}
}

func (w workflowRun[T]) UnderlyingWorkflowRun() client.WorkflowRun {
	return w.wfr
}

func NewWorkflowRunWithErr[T any](run client.WorkflowRun, err error) (WorkflowRun[T], error) {
	return NewWorkflowRun[T](run), err
}
