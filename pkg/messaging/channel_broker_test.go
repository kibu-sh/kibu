package messaging

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

// compile time check that ChannelBroker implements Broker
var _ Broker[Event[any]] = (*ChannelBroker[Event[any]])(nil)

func TestChannelBroker(t *testing.T) {
	ctx := context.Background()

	t.Run("should publish a message", func(t *testing.T) {
		topic := "test"
		expected := "hello world"
		broker := NewChannelBroker[Event[string]]()

		stream, err := broker.Subscribe(ctx, topic)
		require.NoError(t, err)

		err = broker.Publish(ctx, topic, Event[string]{
			ID:   uuid.New(),
			Data: expected,
		})
		require.NoError(t, err)

		actual := <-stream
		require.Equal(t, expected, actual.Data)
	})

	t.Run("should return an error when channel limit is reached", func(t *testing.T) {
		topic := "test"
		broker := NewChannelBroker[Event[string]]()

		broker.ChannelSize = 0
		expected := "hello world"
		err := broker.Publish(ctx, topic, Event[string]{
			ID:   uuid.New(),
			Data: expected,
		})
		require.ErrorIs(t, err, ErrSendFailed)
	})
}
