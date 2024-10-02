package temporal

import (
	"go.temporal.io/sdk/client"
)

type UpdateOptionsProvider interface {
	UpdateOptions(builder UpdateOptionsBuilder) UpdateOptionsBuilder
}

type UpdateOptionFunc func(UpdateOptionsBuilder) UpdateOptionsBuilder

// UpdateOptionsBuilder helps build workflow.ActivityOptions.
type UpdateOptionsBuilder struct {
	updateId            string
	workflowID          string
	runID               string
	updateName          string
	args                []any
	waitForStage        client.WorkflowUpdateStage
	firstExecutionRunID string
}

// NewUpdateOptionsBuilder creates a new UpdateOptionsBuilder.
func NewUpdateOptionsBuilder() UpdateOptionsBuilder {
	return UpdateOptionsBuilder{}
}

func (b UpdateOptionsBuilder) WithUpdateID(id string) UpdateOptionsBuilder {
	b.updateId = id
	return b
}

func (b UpdateOptionsBuilder) WithWorkflowID(id string) UpdateOptionsBuilder {
	b.workflowID = id
	return b
}

func (b UpdateOptionsBuilder) WithRunID(id string) UpdateOptionsBuilder {
	b.runID = id
	return b
}

func (b UpdateOptionsBuilder) WithUpdateName(name string) UpdateOptionsBuilder {
	b.updateName = name
	return b
}

func (b UpdateOptionsBuilder) WithArgs(args ...any) UpdateOptionsBuilder {
	b.args = args
	return b
}

func (b UpdateOptionsBuilder) WithWaitForStage(stage client.WorkflowUpdateStage) UpdateOptionsBuilder {
	b.waitForStage = stage
	return b
}

func (b UpdateOptionsBuilder) WithFirstExecutionRunID(id string) UpdateOptionsBuilder {
	b.firstExecutionRunID = id
	return b
}

func (b UpdateOptionsBuilder) WithOptions(opts ...UpdateOptionFunc) UpdateOptionsBuilder {
	for _, f := range opts {
		b = f(b)
	}
	return b
}

func (b UpdateOptionsBuilder) WithProviders(providers ...UpdateOptionsProvider) UpdateOptionsBuilder {
	for _, p := range providers {
		b = p.UpdateOptions(b)
	}
	return b
}

func (b UpdateOptionsBuilder) WithProvidersWhenSupported(providers ...any) UpdateOptionsBuilder {
	for _, p := range providers {
		if p, ok := p.(UpdateOptionsProvider); ok {
			b = p.UpdateOptions(b)
		}
	}
	return b
}

// Build constructs the workflow.ActivityOptions.
func (b UpdateOptionsBuilder) Build() client.UpdateWorkflowOptions {
	return client.UpdateWorkflowOptions{
		UpdateID:            b.updateId,
		WorkflowID:          b.workflowID,
		RunID:               b.runID,
		UpdateName:          b.updateName,
		Args:                b.args,
		WaitForStage:        b.waitForStage,
		FirstExecutionRunID: b.firstExecutionRunID,
	}
}
