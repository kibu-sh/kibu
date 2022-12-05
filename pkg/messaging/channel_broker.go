package messaging

import (
	"context"
	"github.com/pkg/errors"
	"sync"
)

type ChannelBroker[T any] struct {
	topics      *sync.Map
	ChannelSize int
}

var ErrSendFailed = errors.New("send failed")
var ErrChannelFull = errors.Wrap(ErrSendFailed, "channel is full")

// Publish sends a message to the given topic.
// If the topic doesn't exist it and a matching channel will be created.
func (c ChannelBroker[T]) Publish(ctx context.Context, topic string, message T) (err error) {
	dst, _ := c.topics.LoadOrStore(topic, make(chan T, c.ChannelSize))
	stream := any(dst).(chan T)
	select {
	case <-ctx.Done():
		err = ctx.Err()
		break
	case stream <- message:
		break
	default:
		err = ErrChannelFull
	}
	return
}

// Subscribe returns a channel that will receive messages published to the given topic.
// WARNING: This is intended for testing only
// This ChannelBroker does not support multiple subscribers
// Messages will be consumed by the first subscriber
// If multiple subscribers receive the stream messages will be load balanced by Go channel consumption semantics
func (c ChannelBroker[T]) Subscribe(ctx context.Context, topic string) (stream <-chan T, err error) {
	out, _ := c.topics.LoadOrStore(topic, make(chan T, c.ChannelSize))
	stream = any(out).(chan T)
	return
}

// NewChannelBroker returns a new instance of ChannelBroker
// WARNING: This is intended for testing only
// This ChannelBroker does not support multiple subscribers
func NewChannelBroker[T any]() *ChannelBroker[T] {
	return &ChannelBroker[T]{
		topics:      new(sync.Map),
		ChannelSize: 1000,
	}
}
