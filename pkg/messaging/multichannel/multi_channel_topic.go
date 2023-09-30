package multichannel

import (
	"context"
	"github.com/discernhq/devx/pkg/messaging"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"sync"
)

var (
	ErrSendFailed = errors.New("send failed")
	// TODO: implement a functional pattern for changing the message publishing behavior
	// In some cases we may want messages to be dropped, other times we may want to block
	//ErrChannelFull = errors.Wrap(ErrSendFailed, "channel is full")
)

var _ messaging.Topic[any] = (*Topic[any])(nil)

type Topic[T any] struct {
	cloneFunc      messaging.MessageCloneFunc[T]
	subscribers    sync.Map
	subChannelSize int
}

// NewTopic returns a new instance of messaging.Topic
func NewTopic[T any](clone messaging.MessageCloneFunc[T], subChannelSize int) *Topic[T] {
	return &Topic[T]{
		cloneFunc:      clone,
		subChannelSize: subChannelSize,
	}
}

func NewTopicWithDefaults[T any]() *Topic[T] {
	return NewTopic[T](messaging.PassThoughClone[T], 0)
}

// Publish sends a message to all subscribers of this topic in parallel
// All subscribers must be able to receive the message or this function will block
func (t *Topic[T]) Publish(ctx context.Context, message T) (err error) {
	eg, ctx := errgroup.WithContext(ctx)

	t.subscribers.Range(func(key, value any) bool {
		eg.Go(func() error {
			return t.publishSingle(value.(chan T), message, ctx)
		})
		return true
	})

	err = eg.Wait()

	return
}

func (t *Topic[T]) publishSingle(sub chan T, message T, ctx context.Context) error {
	clone, err := t.cloneFunc(message)
	if err != nil {
		return err
	}

	select {
	case sub <- clone:
	case <-ctx.Done():
		err = ctx.Err()
	}

	return err
}

// Subscribe returns a channel that will receive messages published to this topic
func (t *Topic[T]) Subscribe(_ context.Context) (stream messaging.Stream[T], err error) {
	out := make(chan T, t.subChannelSize)
	t.subscribers.Store(out, out)
	stream = messaging.NewChannelStream(out, func() {
		t.subscribers.Delete(out)
	})

	return
}
