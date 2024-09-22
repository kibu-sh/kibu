package temporal

import (
	"context"
	"github.com/pkg/errors"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"time"
)

type FutureCallback[T any] func(Future[T])

type Future[T any] interface {
	Get(ctx workflow.Context) (res T, err error)
	IsReady() bool
	Underlying() workflow.Future
	Select(workflow.Selector, FutureCallback[T]) workflow.Selector
}

type ChildWorkflowFuture[T any] interface {
	Future[T]
	UnderlyingChildWorkflowFuture() workflow.ChildWorkflowFuture
	GetChildWorkflowExecution() Future[workflow.Execution]
}

type WorkflowRun[T any] interface {
	GetID() string
	GetRunID() string
	Underlying() client.WorkflowRun
	Get(ctx context.Context) (result T, err error)
	GetWithOptions(ctx context.Context, options client.WorkflowRunGetOptions) (result T, err error)
}

func WithDefaultActivityOptions(ctx workflow.Context) workflow.Context {
	return workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 30,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 5,
		},
	})
}

func WithInfiniteRetryActivityPolicy(ctx workflow.Context) workflow.Context {
	return workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 30,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	})
}

func WithInfiniteRetryAndMaxBackoffActivityPolicy(ctx workflow.Context, maxInterval time.Duration) workflow.Context {
	return workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 30,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    0,
			BackoffCoefficient: 2,
			MaximumInterval:    maxInterval,
		},
	})
}

func WithChildWorkflowParentClosePolicy_ABANDON(ctx workflow.Context) workflow.Context {
	return workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON,
	})
}

func ErrorIs(err, target error) (match bool) {
	var receivedErr *temporal.ApplicationError
	var targetErr *temporal.ApplicationError
	if errors.As(err, &receivedErr) && errors.As(target, &targetErr) {
		return receivedErr.Type() == targetErr.Type()
	}
	return false
}
