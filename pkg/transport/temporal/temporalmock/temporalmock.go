package temporalmock

import (
	"github.com/kibu-sh/kibu/pkg/transport/temporal"
	"go.temporal.io/sdk/workflow"
)

var _ temporal.Future[any] = Future[any]{}

type Future[T any] struct {
	Result T
	Ready  bool
}

func (f Future[T]) Select(selector workflow.Selector, f2 temporal.FutureCallback[T]) workflow.Selector {
	return selector.AddFuture(f.Underlying(), func(workflow.Future) {
		if f2 != nil {
			f2(f)
		}
	})
}

func (f Future[T]) Underlying() workflow.Future {
	panic("this is a mock and doesn't support method Underlying")
}

func (f Future[T]) Get(ctx workflow.Context) (res T, err error) {
	return f.Result, nil
}

func (f Future[T]) IsReady() bool {
	return f.Ready
}

func NewFuture[T any](data T, ready bool) temporal.Future[T] {
	return Future[T]{Result: data, Ready: ready}
}
