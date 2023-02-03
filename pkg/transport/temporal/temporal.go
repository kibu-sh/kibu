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

type Workflow struct {
	Name    string
	Handler any
}

type Activity struct {
	Name    string
	Handler any
}

type WorkflowFactory interface {
	WorkflowFactory() []*Workflow
}

type ActivityFactory interface {
	ActivityFactory() []*Activity
}

func NewWorkflow(name string, handler any) *Workflow {
	return &Workflow{
		Name:    name,
		Handler: handler,
	}
}

func NewActivity(name string, handler any) *Activity {
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
		fmt.Printf("registering workflow %s\n", wf.Name)
		w.RegisterWorkflowWithOptions(wf.Handler, workflow.RegisterOptions{
			Name: wf.Name,
		})
	}
	for _, act := range activities {
		fmt.Printf("registering activity %s\n", act.Name)
		w.RegisterActivityWithOptions(act.Handler, activity.RegisterOptions{
			Name: act.Name,
		})
	}
	return
}

type RetryPolicy = temporal.RetryPolicy
