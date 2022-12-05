package database

import "context"

type Model interface {
	PrimaryKey() string
}

type Repository[T Model] interface {
	Save(ctx context.Context, model T) (err error)
	Delete(ctx context.Context, model T) (err error)
	FindOne(ctx context.Context, primaryKey string) (model *T, err error)
	FindOneOrThrow(ctx context.Context, primaryKey string) (model *T, err error)
}
