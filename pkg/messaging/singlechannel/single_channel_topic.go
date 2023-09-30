package singlechannel

import (
	"context"
	"github.com/pkg/errors"
)

type Topic[T any] struct {
	maxSize int
	dest    chan T
}

var ErrSendFailed = errors.New("send failed")
var ErrChannelFull = errors.Wrap(ErrSendFailed, "channel is full")

// Publish sends a message to the given topic.
// If the topic doesn't exist it and a matching channel will be created.
func (c *Topic[T]) Publish(ctx context.Context, message T) (err error) {
	select {
	case <-ctx.Done():
		err = ctx.Err()
		break
	case c.dest <- message:
		break
	default:
		err = ErrChannelFull
	}
	return
}

// Subscribe returns a channel that will receive messages published to the given topic.
// WARNING: This is intended for testing only
// This Topic does not support multiple subscribers
// Messages will be consumed by the first subscriber
// If multiple subscribers receive the stream messages will be load balanced by Go channel consumption semantics
func (c *Topic[T]) Subscribe(_ context.Context) (stream <-chan T, err error) {
	if c.dest == nil {
		c.dest = make(chan T, c.maxSize)
	}
	stream = c.dest
	return
}

func (c *Topic[T]) Unsubscribe(_ context.Context, stream <-chan T) (err error) {
	if c.dest != stream {
		err = errors.New("cannot unsubscribe from a stream that was not returned by Subscribe")
		return
	}

	close(c.dest)
	c.dest = nil
	return
}

// NewTopic returns a new instance of Topic
// WARNING: This is intended for testing only
// This Topic does not support multiple subscribers
func NewTopic[T any]() *Topic[T] {
	return &Topic[T]{
		maxSize: 1000,
	}
}
