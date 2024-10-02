package temporal

import (
	"github.com/pkg/errors"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
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

type ExecuteParams[T any] struct {
	Request T
	Options []WorkflowOptionFunc
}

type ExecuteWithSignalParams[T, S any] struct {
	Request T
	Signal  S
	Options []WorkflowOptionFunc
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

type WorkerFactory interface {
	Build() worker.Worker
}
