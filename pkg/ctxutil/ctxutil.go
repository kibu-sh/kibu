package ctxutil

import (
	"context"
	"github.com/pkg/errors"
)

var ErrNotFoundInContext = errors.New("not found in context")

type Loader[T any] interface {
	Load(ctx context.Context) (T, error)
}

type Saver[T any] interface {
	Save(ctx context.Context, v T) context.Context
}

type Provider[T any] interface {
	Loader[T]
	Saver[T]
}

type LoaderFunc[T any] func(ctx context.Context) (T, error)

var _ Provider[any] = (*Store[any, any])(nil)

type Store[T any, K any] struct {
	key *K
}

func (s *Store[T, K]) Save(ctx context.Context, v T) context.Context {
	return context.WithValue(ctx, s.key, v)
}

func (s *Store[T, K]) Load(ctx context.Context) (r T, err error) {
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
