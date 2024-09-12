package singlechannel

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/messaging"
	"github.com/pkg/errors"
)

var defaultMaxSize = 1000

// compile time check that Topic implements Broker
var _ messaging.Topic[any] = (*Topic[any])(nil)
var _ messaging.Stream[any] = (*ChannelStream[any])(nil)

type Topic[T any] struct {
	maxSize int
	dest    *ChannelStream[T]
}

type ChannelStream[T any] struct {
	ch chan T
}

func (c ChannelStream[T]) Channel() <-chan T {
	return c.ch
}

func (c ChannelStream[T]) Next(ctx context.Context) (message T, hasNext bool, err error) {
	select {
	case <-ctx.Done():
		err = ctx.Err()
		return
	case message, hasNext = <-c.ch:
		return
	}
}

func (c ChannelStream[T]) Unsubscribe() {
	close(c.ch)
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
	case c.dest.ch <- message:
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
func (c *Topic[T]) Subscribe(_ context.Context) (messaging.Stream[T], error) {
	return c.dest, nil
}

// NewTopic returns a new instance of Topic
// WARNING: This is intended for testing only
// This Topic does not support multiple subscribers
func NewTopic[T any]() *Topic[T] {
	return &Topic[T]{
		maxSize: defaultMaxSize,
		dest:    newChannelStream[T](defaultMaxSize),
	}
}

func newChannelStream[T any](maxSize int) *ChannelStream[T] {
	return &ChannelStream[T]{ch: make(chan T, maxSize)}
}
