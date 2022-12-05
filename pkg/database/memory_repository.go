package database

import (
	"context"
	"github.com/discernhq/devx/pkg/utils"
	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("not found")

type InMemoryRepository[T Model] struct {
	objects *utils.SyncMap[T]
}

func (t *InMemoryRepository[T]) Save(ctx context.Context, model T) (err error) {
	t.objects.Store(model.PrimaryKey(), &model)
	return
}

func (t *InMemoryRepository[T]) Delete(ctx context.Context, model T) (err error) {
	t.objects.Delete(model.PrimaryKey())
	return
}

func (t *InMemoryRepository[T]) FindOne(ctx context.Context, primaryKey string) (model *T, err error) {
	model, err = t.FindOneOrThrow(ctx, primaryKey)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	return
}

func (t *InMemoryRepository[T]) FindOneOrThrow(ctx context.Context, primaryKey string) (model *T, err error) {
	model, ok := t.objects.Load(primaryKey)
	if !ok {
		err = errors.Wrapf(ErrNotFound, "%T by primary key %s", model, primaryKey)
		return
	}
	return
}

func NewMemoryRepository[T Model]() *InMemoryRepository[T] {
	return &InMemoryRepository[T]{
		objects: utils.NewSyncMap[T](),
	}
}
