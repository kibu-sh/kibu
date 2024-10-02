package messaging

import (
	"context"
	"github.com/google/uuid"
)

type Publisher[T any] interface {
	Publish(ctx context.Context, message T) (err error)
}

type Stream[T any] interface {
	Unsubscribe()
	Channel() (stream <-chan T)
	Next(ctx context.Context) (message T, hasNext bool, err error)
}

type Consumer[T any] interface {
	Subscribe(ctx context.Context) (stream Stream[T], err error)
}

type Topic[T any] interface {
	Publisher[T]
	Consumer[T]
}

type Broker[T any] interface {
	Topic(string) (Topic[T], error)
}

type Event[T any] struct {
	ID   uuid.UUID
	Data T
}

func NewEvent[T any](data T) Event[T] {
	return Event[T]{
		ID:   uuid.New(),
		Data: data,
	}
}

type MessageCloneFunc[T any] func(T) (T, error)

func PassThoughClone[T any](t T) (T, error) {
	return t, nil
}

var _ Stream[any] = (*ChannelStream[any])(nil)

type ChannelStream[T any] struct {
	channel chan T
	cleanup func()
}

func (c ChannelStream[T]) Channel() (stream <-chan T) {
	return c.channel
}

func (c ChannelStream[T]) Next(ctx context.Context) (message T, hasNext bool, err error) {
	select {
	case <-ctx.Done():
		err = ctx.Err()
		return
	case message, hasNext = <-c.channel:
		return
	}
}

func (c ChannelStream[T]) Unsubscribe() {
	close(c.channel)
	c.cleanup()
}

func NewChannelStream[T any](channel chan T, cleanup func()) *ChannelStream[T] {
	return &ChannelStream[T]{
		channel: channel,
		cleanup: cleanup,
	}
}
