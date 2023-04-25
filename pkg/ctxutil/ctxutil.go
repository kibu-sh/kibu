package ctxutil

import (
	"context"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

var ErrNotFoundInContext = errors.New("not found in context")

type Loader[T any] interface {
	Load(ctx ValueContainer) (T, error)
}

type Saver[T any] interface {
	Save(ctx context.Context, v T) context.Context
}

type WorkflowSaver[T any] interface {
	SaveToWorkflow(ctx workflow.Context, v T) workflow.Context
}

type Provider[T any] interface {
	Loader[T]
	Saver[T]
	WorkflowSaver[T]
}

type LoaderFunc[T any] func(ctx ValueContainer) (T, error)

var _ Provider[any] = (*Store[any, any])(nil)

type Store[T any, K any] struct {
	key *K
}

func (s *Store[T, K]) Save(ctx context.Context, v T) context.Context {
	return context.WithValue(ctx, s.key, v)
}

func (s *Store[T, K]) SaveToWorkflow(ctx workflow.Context, v T) workflow.Context {
	return workflow.WithValue(ctx, s.key, v)
}

type ValueContainer interface {
	Value(any) any
}

func (s *Store[T, K]) Load(ctx ValueContainer) (r T, err error) {
	v := ctx.Value(s.key)
	if v == nil {
		err = errors.Wrapf(
			ErrNotFoundInContext,
			"cannot find %T by key %T", new(T), s.key,
		)
		return
	}

	r = v.(T)
	return
}

func NewStore[T any, K any]() *Store[T, K] {
	return &Store[T, K]{
		key: new(K),
	}
}

// ensures that WorkflowPropagator implements the temporal.ContextPropagator interface
var _ workflow.ContextPropagator = (*WorkflowPropagator[any])(nil)

type Propagator[T any] interface {
	workflow.ContextPropagator
}

type WorkflowPropagator[T any] struct {
	provider Provider[T]
	key      string
}

func NewPropagator[T any](key string, provider Provider[T]) *WorkflowPropagator[T] {
	return &WorkflowPropagator[T]{
		key:      key,
		provider: provider,
	}
}

// Inject injects information from a Go Context into headers
func (p *WorkflowPropagator[T]) Inject(ctx context.Context, writer workflow.HeaderWriter) (err error) {
	value, err := p.provider.Load(ctx)
	if errors.Is(err, ErrNotFoundInContext) {
		return nil
	}

	if err != nil {
		return
	}

	payload, err := converter.GetDefaultDataConverter().ToPayload(value)
	if err != nil {
		return
	}

	writer.Set(p.key, payload)
	return
}

// InjectFromWorkflow injects information from workflow context into headers
func (p *WorkflowPropagator[T]) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) (err error) {
	value, err := p.provider.Load(ctx)
	if errors.Is(err, ErrNotFoundInContext) {
		return nil
	}

	if err != nil {
		return
	}

	payload, err := converter.GetDefaultDataConverter().ToPayload(value)
	if err != nil {
		return
	}

	writer.Set(p.key, payload)
	return
}

// Extract extracts context information from headers and returns a context
func (p *WorkflowPropagator[T]) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	payload, ok := reader.Get(p.key)
	if !ok {
		return ctx, nil
	}

	value := new(T)
	if err := converter.GetDefaultDataConverter().FromPayload(payload, value); err != nil {
		return ctx, nil
	}
	return p.provider.Save(ctx, *value), nil
}

// ExtractToWorkflow extracts context information from headers and returns a workflow context
func (p *WorkflowPropagator[T]) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	payload, ok := reader.Get(p.key)
	if !ok {
		return ctx, nil
	}

	value := new(T)
	if err := converter.GetDefaultDataConverter().FromPayload(payload, value); err != nil {
		return ctx, nil
	}
	return p.provider.SaveToWorkflow(ctx, *value), nil
}
