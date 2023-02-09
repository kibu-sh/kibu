package ctxutil

import (
	"context"
	"github.com/pkg/errors"
)

var ErrNotFoundInContext = errors.New("not found in context")

type Store[T any, K any] struct {
	key *K
}

func (s *Store[T, K]) Save(ctx context.Context, v *T) context.Context {
	return context.WithValue(ctx, s.key, v)
}

func (s *Store[T, K]) Load(ctx context.Context) (*T, error) {
	v := ctx.Value(s.key)
	if v == nil {
		return nil, errors.Wrapf(
			ErrNotFoundInContext,
			"cannot find %T by key %T", new(T), s.key,
		)
	}

	return v.(*T), nil
}

func NewStore[T any, K any]() *Store[T, K] {
	return &Store[T, K]{
		key: new(K),
	}
}
