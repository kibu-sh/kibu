package temporal

import (
	"context"
	"go.temporal.io/sdk/client"
)

type GetHandleOpts struct {
	WorkflowID string
	RunID      string
}

type UpdateHandle[T any] interface {
	// UpdateID returns the ID of this update
	UpdateID() string

	// WorkflowID returns the workflowID of the workflow
	WorkflowID() string

	// RunID returns the runID of the workflow
	RunID() string

	// Get blocks until the future is ready
	Get(ctx context.Context) (T, error)
}

var _ UpdateHandle[any] = (*updateHandle[any])(nil)

type updateHandle[T any] struct {
	handle client.WorkflowUpdateHandle
}

func (u updateHandle[T]) UpdateID() string {
	return u.handle.UpdateID()
}

func (u updateHandle[T]) WorkflowID() string {
	return u.handle.WorkflowID()
}

func (u updateHandle[T]) RunID() string {
	return u.handle.RunID()
}

func (u updateHandle[T]) Get(ctx context.Context) (T, error) {
	var result T
	err := u.handle.Get(ctx, &result)
	return result, err
}

func NewUpdateHandle[T any](handle client.WorkflowUpdateHandle) UpdateHandle[T] {
	return &updateHandle[T]{handle}
}
