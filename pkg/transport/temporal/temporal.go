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
			EnableLoggingInReplay: true,
			WorkerStopTimeout:     time.Second * 30,
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

type Future[T any] struct {
	workflow.Future
}

func (f Future[T]) Get(ctx workflow.Context) (res T, err error) {
	err = f.Future.Get(ctx, &res)
	return
}

func NewFuture[T any](f workflow.Future) Future[T] {
	return Future[T]{f}
}

type WorkflowRun[T any] struct {
	client.WorkflowRun
}

func NewWorkflowRun[T any](run client.WorkflowRun) WorkflowRun[T] {
	return WorkflowRun[T]{run}
}

func NewWorkflowRunWithErr[T any](run client.WorkflowRun, err error) (WorkflowRun[T], error) {
	return NewWorkflowRun[T](run), err
}

func (w WorkflowRun[T]) Get(ctx context.Context) (res T, err error) {
	err = w.WorkflowRun.Get(ctx, &res)
	return
}

func (w WorkflowRun[T]) GetWithOptions(ctx context.Context, options client.WorkflowRunGetOptions) (res T, err error) {
	err = w.WorkflowRun.GetWithOptions(ctx, res, options)
	return
}
