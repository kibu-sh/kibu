package temporal

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/messaging"
	"go.temporal.io/sdk/client"
)

var _ messaging.Publisher[any] = (*SignalPublisherClient[any, any])(nil)

type SignalPublisherClient[T any, U any] struct {
	Client               client.Client
	WorkflowID           string
	SignalName           string
	Workflow             any
	WorkflowStartMessage U
	StartOptions         client.StartWorkflowOptions
}

func (s SignalPublisherClient[T, U]) Publish(ctx context.Context, message T) (err error) {
	_, err = s.Client.SignalWithStartWorkflow(
		ctx,
		s.WorkflowID,
		s.SignalName,
		message,
		s.StartOptions,
		s.Workflow,
	)
	return
}
