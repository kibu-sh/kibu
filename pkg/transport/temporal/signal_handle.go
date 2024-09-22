package temporal

import "go.temporal.io/sdk/workflow"

type SignalHandle interface {
	// Get blocks until the future is ready
	Get(workflow.Context) error

	// IsReady returns true when Get is guaranteed not to block
	IsReady() bool
}

var _ SignalHandle = (*signalHandle[any])(nil)

type signalHandle[T any] struct {
	future workflow.Future
}

func (s signalHandle[T]) Get(ctx workflow.Context) error {
	return s.future.Get(ctx, nil)
}

func (s signalHandle[T]) IsReady() bool {
	return s.future.IsReady()
}

func NewSignalHandle[T any](future workflow.Future) SignalHandle {
	return &signalHandle[T]{future}
}
