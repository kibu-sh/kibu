package temporal

import (
	"context"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"time"
)

type TaskQueue string
type WorkflowFunc func(ctx workflow.Context, req any) (res any, err error)

type ActivityFunc func(ctx context.Context, req any) (res any, err error)

type Workflow struct {
	Name    string
	Handler func(ctx workflow.Context, req any) (res any, err error)
}

type Activity struct {
	Name    string
	Handler func(ctx context.Context, req any) (res any, err error)
}
type WorkflowFactory interface {
	WorkflowFactory() []*Workflow
}

type ActivityFactory interface {
	ActivityFactory() []*Activity
}

func NewWorkflow(name string, handler WorkflowFunc) *Workflow {
	return &Workflow{
		Name:    name,
		Handler: handler,
	}
}

func NewActivity(name string, handler ActivityFunc) *Activity {
	return &Activity{
		Name:    name,
		Handler: handler,
	}
}

func NewWorker(
	client client.Client,
	queue TaskQueue,
	workflows []*Workflow,
	activities []*Activity,
) (w worker.Worker, err error) {
	// TODO: pre production tuning
	w = worker.New(client, string(queue), worker.Options{
		EnableLoggingInReplay: true,
		WorkerStopTimeout:     time.Second * 30,
	})
	for _, wf := range workflows {
		w.RegisterWorkflowWithOptions(wf.Handler, workflow.RegisterOptions{
			Name: wf.Name,
		})
	}
	for _, act := range activities {
		w.RegisterActivityWithOptions(act.Handler, activity.RegisterOptions{
			Name: act.Name,
		})
	}
	return
}
