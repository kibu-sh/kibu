package temporal

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
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
	logger zerolog.Logger,
) (workers []worker.Worker, err error) {
	defByTaskQueue := lo.GroupBy(workerDefs, func(def *Worker) string {
		return def.TaskQueue
	})

	for queue, workerDefs := range defByTaskQueue {
		w := worker.New(client, queue, worker.Options{
			EnableLoggingInReplay:       true,
			WorkerStopTimeout:           time.Second * 30,
			DisableRegistrationAliasing: true,
		})

		for _, def := range workerDefs {
			logger.Info().
				Str("queue", queue).
				Str("type", def.Type).
				Str("name", def.Name).
				Msg("registering worker")

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

type RetryPolicy = temporal.RetryPolicy

type Future[T any] interface {
	Get(ctx workflow.Context) (res T, err error)
	IsReady() bool
}

type future[T any] struct {
	wf workflow.Future
}

func (f future[T]) Get(ctx workflow.Context) (res T, err error) {
	err = f.wf.Get(ctx, &res)
	return
}

func (f future[T]) IsReady() bool {
	return f.wf.IsReady()
}

func NewFuture[T any](f workflow.Future) Future[T] {
	return future[T]{f}
}

type WorkflowRun[T any] interface {
	GetID() string
	GetRunID() string
	Get(ctx context.Context) (result T, err error)
	GetWithOptions(ctx context.Context, options client.WorkflowRunGetOptions) (result T, err error)
}

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

func NewWorkflowRunWithErr[T any](run client.WorkflowRun, err error) (WorkflowRun[T], error) {
	return NewWorkflowRun[T](run), err
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
