package ctxutil

import (
	"context"
	"github.com/pkg/errors"
)

var ErrNotFoundInContext = errors.New("not found in context")

type Store[T any] struct {
	key any
}

func (s *Store[T]) Save(ctx context.Context, v *T) context.Context {
	return context.WithValue(ctx, s.key, v)
}

func (s *Store[T]) Load(ctx context.Context) (*T, error) {
	v := ctx.Value(s.key)
	if v == nil {
		return nil, errors.Wrapf(
			ErrNotFoundInContext,
			"cannot find %T by key %T", new(T), s.key,
		)
	}

	return v.(*T), nil
}

func NewStore[T any](key any) *Store[T] {
	return &Store[T]{
		key: key,
	}
}
