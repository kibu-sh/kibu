package singlechannel

import (
	"context"
	"github.com/discernhq/devx/pkg/messaging"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTopic(t *testing.T) {
	ctx := context.Background()
	type stringEvent messaging.Event[string]

	t.Run("should publish a message", func(t *testing.T) {
		expected := "hello world"
		broker := NewTopic[stringEvent]()

		stream, err := broker.Subscribe(ctx)
		require.NoError(t, err)

		err = broker.Publish(ctx, stringEvent{
			ID:   uuid.New(),
			Data: expected,
		})
		require.NoError(t, err)

		actual := <-stream.Channel()
		require.Equal(t, expected, actual.Data)
	})

	t.Run("should return an error when channel limit is reached", func(t *testing.T) {
		broker := NewTopic[stringEvent]()
		broker.dest = newChannelStream[stringEvent](0)
		expected := "hello world"
		err := broker.Publish(ctx, stringEvent{
			ID:   uuid.New(),
			Data: expected,
		})
		require.ErrorIs(t, err, ErrSendFailed)
	})
}
