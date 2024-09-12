package temporalmock

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/transport/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

var _ temporal.Future[any] = Future[any]{}
var _ temporal.WorkflowRun[any] = WorkflowRun[any]{}

type Future[T any] struct {
	Result T
	Ready  bool
}

func (f Future[T]) UnderlyingFuture() workflow.Future {
	panic("this is a mock and doesn't support method UnderlyingFuture")
}

func (f Future[T]) Get(ctx workflow.Context) (res T, err error) {
	return f.Result, nil
}

func (f Future[T]) IsReady() bool {
	return f.Ready
}

func NewFuture[T any](data T, ready bool) temporal.Future[T] {
	return Future[T]{Result: data, Ready: ready}
}

type WorkflowRun[T any] struct {
	ID     string
	RunID  string
	Result T
}

func (w WorkflowRun[T]) UnderlyingWorkflowRun() client.WorkflowRun {
	panic("this is a mock and doesn't support method UnderlyingWorkflowRun")
}

func (w WorkflowRun[T]) GetID() string {
	return w.ID
}

func (w WorkflowRun[T]) GetRunID() string {
	return w.RunID
}

func (w WorkflowRun[T]) Get(ctx context.Context) (result T, err error) {
	return w.Result, nil
}

func (w WorkflowRun[T]) GetWithOptions(ctx context.Context, options client.WorkflowRunGetOptions) (result T, err error) {
	return w.Result, nil
}

func NewWorkflowRun[T any](id, runID string, data T) temporal.WorkflowRun[T] {
	return WorkflowRun[T]{ID: id, RunID: runID, Result: data}
}
