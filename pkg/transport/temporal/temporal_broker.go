package temporal

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/messaging"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"time"
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

type SignalChannel[T any] struct {
	ch workflow.ReceiveChannel
}

func (s SignalChannel[T]) Len() int {
	return s.ch.Len()
}

func (s SignalChannel[T]) Receive(ctx workflow.Context) (message T, more bool) {
	more = s.ch.Receive(ctx, &message)
	return
}

func (s SignalChannel[T]) ReceiveAsync() (message T, ok bool) {
	ok = s.ch.ReceiveAsync(&message)
	return
}

func (s SignalChannel[T]) ReceiveWithTimeout(ctx workflow.Context, timeout time.Duration) (message T, ok, more bool) {
	more, ok = s.ch.ReceiveWithTimeout(ctx, timeout, &message)
	return
}

func (s SignalChannel[T]) ReceiveAsyncWithMoreFlag() (message T, ok bool, more bool) {
	more, ok = s.ch.ReceiveAsyncWithMoreFlag(&message)
	return
}

func NewSignalChannel[T any](ctx workflow.Context, signal string) SignalChannel[T] {
	return SignalChannel[T]{ch: workflow.GetSignalChannel(ctx, signal)}
}
