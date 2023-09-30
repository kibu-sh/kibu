package multichannel

import (
	"context"
	"github.com/discernhq/devx/pkg/messaging"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type anyMessage = messaging.Event[any]
type stringMessage = messaging.Event[string]

// compile time check that Broker implements Broker
var _ messaging.Topic[anyMessage] = (*Topic[anyMessage])(nil)

func newTestTopic() *Topic[stringMessage] {
	return NewTopicWithDefaults[stringMessage]()
}

func TestTopic(t *testing.T) {
	t.Run("should publish a message", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		expected := "hello world"
		topic := newTestTopic()
		stream, err := topic.Subscribe(ctx)
		require.NoError(t, err)

		go func() {
			err = topic.Publish(ctx, stringMessage{
				ID:   uuid.New(),
				Data: expected,
			})
			require.NoError(t, err)
		}()

		actual, _, err := stream.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, expected, actual.Data)
	})

	// TODO: bring this test back when we have implemented the functional pattern for
	// changing the publish strategy
	//t.Run("should return an error when channel limit is reached", func(t *testing.T) {
	//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	//	defer cancel()
	//	topic := newTestTopic()
	//	topic.subChannelSize = 0
	//	stream, err := topic.Subscribe(ctx)
	//	require.NoError(t, err)
	//	require.Len(t, stream.Channel(), 0)
	//
	//	expected := "hello world"
	//	err = topic.Publish(ctx, stringMessage{
	//		ID:   uuid.New(),
	//		Data: expected,
	//	})
	//	require.ErrorIs(t, err, ErrSendFailed)
	//})

	t.Run("should deliver multiple messages", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		topic := newTestTopic()
		sub1, err := topic.Subscribe(ctx)
		require.NoError(t, err)

		sub2, err := topic.Subscribe(ctx)
		require.NoError(t, err)

		expected := "hello world"

		go func() {
			err = topic.Publish(ctx, stringMessage{
				ID:   uuid.New(),
				Data: expected,
			})
			require.NoError(t, err)
		}()

		m1 := <-sub1.Channel()
		m2 := <-sub2.Channel()
		require.Equal(t, expected, m1.Data)
		require.Equal(t, expected, m2.Data)
		require.Equal(t, m1.ID, m2.ID)
		require.Equal(t, m1.Data, m2.Data)
	})

	t.Run("should unsubscribe and close channels", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		topic := newTestTopic()
		sub1, err := topic.Subscribe(ctx)
		require.NoError(t, err)
		sub1.Unsubscribe()

		sub2, err := topic.Subscribe(ctx)
		require.NoError(t, err)
		sub2.Unsubscribe()

		require.NoError(t, err)

		_, s1Open := <-sub1.Channel()
		_, s2Open := <-sub2.Channel()
		require.Falsef(t, s1Open, "sub1 should be closed")
		require.Falsef(t, s2Open, "sub2 should be closed")
	})
}
