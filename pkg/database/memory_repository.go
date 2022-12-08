package database

import (
	"context"
	"github.com/discernhq/devx/pkg/utils"
	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("not found")

type Model interface {
	PrimaryKey() string
}

type MemoryRepository[T Model] struct {
	objects *utils.SyncMap[T]
}

func asModel(v any) Model {
	return v.(Model)
}

func (t *MemoryRepository[T]) Save(ctx context.Context, model *T) (err error) {
	t.objects.Store(asModel(model).PrimaryKey(), model)
	return
}

func (t *MemoryRepository[T]) Delete(ctx context.Context, model *T) (err error) {
	t.objects.Delete(asModel(model).PrimaryKey())
	return
}

func (t *MemoryRepository[T]) FindOne(ctx context.Context, primaryKey string) (model *T, err error) {
	model, err = t.FindOneOrThrow(ctx, primaryKey)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	return
}

func (t *MemoryRepository[T]) FindOneOrThrow(ctx context.Context, primaryKey string) (model *T, err error) {
	model, ok := t.objects.Load(primaryKey)
	if !ok {
		err = errors.Wrapf(ErrNotFound, "%T by primary key %s", model, primaryKey)
		return
	}
	return
}

func NewMemoryRepository[T Model]() *MemoryRepository[T] {
	return &MemoryRepository[T]{
		objects: utils.NewSyncMap[T](),
	}
}
