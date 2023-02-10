package temporal

import (
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
