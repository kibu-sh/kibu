package temporal

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"log/slog"
	"time"
)

type TaskQueue string
type Worker struct {
	Name      string
	TaskQueue string
	Type      string
	Handler   any
}

type WorkerFactory interface {
	WorkerFactory() []*Worker
}

func NewWorker(
	client client.Client,
	workerDefs []*Worker,
	logger *slog.Logger,
) (workers []worker.Worker, err error) {
	defByTaskQueue := lo.GroupBy(workerDefs, func(def *Worker) string {
		return def.TaskQueue
	})

	for queue, workerDefs := range defByTaskQueue {
		w := worker.New(client, queue, worker.Options{
			EnableLoggingInReplay:            true,
			WorkerStopTimeout:                time.Second * 30,
			DisableRegistrationAliasing:      true,
			MaxConcurrentActivityTaskPollers: 4,
			MaxConcurrentWorkflowTaskPollers: 4,
		})

		for _, def := range workerDefs {
			logger.With("queue", queue).
				With("type", def.Type).
				With("name", def.Name).
				Debug(fmt.Sprintf("[kibu.transport.temporal] %s %s", def.Type, def.Name))
			switch def.Type {
			case "workflow":
				w.RegisterWorkflowWithOptions(def.Handler, workflow.RegisterOptions{
					Name: def.Name,
				})
			case "activity":
				w.RegisterActivityWithOptions(def.Handler, activity.RegisterOptions{
					Name: def.Name,
				})
			}
		}

		workers = append(workers, w)
	}
	return
}

type Future[T any] interface {
	Get(ctx workflow.Context) (res T, err error)
	IsReady() bool
	UnderlyingFuture() workflow.Future
}

type ChildWorkflowFuture[T any] interface {
	Future[T]
	UnderlyingChildWorkflowFuture() workflow.ChildWorkflowFuture
	GetChildWorkflowExecution() Future[T]
}

type WorkflowRun[T any] interface {
	GetID() string
	GetRunID() string
	Get(ctx context.Context) (result T, err error)
	GetWithOptions(ctx context.Context, options client.WorkflowRunGetOptions) (result T, err error)
	UnderlyingWorkflowRun() client.WorkflowRun
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

type WorkflowOptionsCtor interface {
	WorkflowOptionsCtor(ctx context.Context, opts StartWorkflowOptions) (StartWorkflowOptions, error)
}

type ChildWorkflowOptionsCtor interface {
	ChildWorkflowOptionsCtor(ctx workflow.Context, opts ChildWorkflowOptions) (ChildWorkflowOptions, error)
}

type ActivityOptionsCtor interface {
	ActivityOptionsCtor(ctx workflow.Context, opts ActivityOptions) ActivityOptions
}

type ExecuteWorkflowOptions struct {
	Client      client.Client
	WorkflowRef string
	Options     StartWorkflowOptions
}

func ExecuteWorkflow[Request, Response any](ctx context.Context, req Request, params ExecuteWorkflowOptions) (WorkflowRun[Response], error) {
	opts := client.StartWorkflowOptions{
		ID:        uuid.New().String(),
		TaskQueue: "default",
	}

	if override, ok := any(req).(WorkflowOptionsCtor); ok {
		opts = override.WorkflowStartOptions(opts)
	}

	return NewWorkflowRunWithErr[Res](
		params.Client.ExecuteWorkflow(ctx, opts, "exp.Workflow.Start", req),
	)
}

type SideEffectFunc[T any] func(ctx workflow.Context) T

func SideEffect[T any](ctx workflow.Context, fun SideEffectFunc[T]) (result T, err error) {
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) any {
		return fun(ctx)
	}).Get(&result)
	return
}

func NewWorkflowID(ctx workflow.Context) (string, error) {
	return SideEffect(ctx, func(ctx workflow.Context) string {
		return uuid.New().String()
	})
}
