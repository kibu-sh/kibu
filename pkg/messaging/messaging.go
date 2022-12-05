package messaging

import (
	"context"
	"github.com/google/uuid"
)

type Publisher[T any] interface {
	Publish(ctx context.Context, topic string, message T) (err error)
}

type Consumer[T any] interface {
	Subscribe(ctx context.Context, topic string) (stream <-chan T, err error)
}

type Broker[T any] interface {
	Publisher[T]
	Consumer[T]
}

type Event[T any] struct {
	ID   uuid.UUID
	Data T
}
