package temporal

import (
	"fmt"
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
	WorkerFactory() []*WorkerFactory
}

func NewWorker(
	client client.Client,
	workerDefs []*Worker,
) (workers []worker.Worker, err error) {
	w := worker.New(client, "default", worker.Options{
		Identity:              "my-worker",
		EnableLoggingInReplay: true,
		WorkerStopTimeout:     time.Second * 30,
	})
	for _, def := range workerDefs {
		// TODO: pre production tuning
		fmt.Printf("registering %s %s\n", def.Type, def.Name)

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
	return
}

type RetryPolicy = temporal.RetryPolicy
