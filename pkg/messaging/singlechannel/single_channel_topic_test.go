package singlechannel

import (
	"context"
	"github.com/discernhq/devx/pkg/messaging"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

type anyEvent messaging.Event[any]
type stringEvent messaging.Event[string]

// compile time check that Topic implements Broker
var _ messaging.Topic[anyEvent] = (*Topic[anyEvent])(nil)

func TestTopic(t *testing.T) {
	ctx := context.Background()

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

		actual := <-stream
		require.Equal(t, expected, actual.Data)
	})

	t.Run("should return an error when channel limit is reached", func(t *testing.T) {
		broker := NewTopic[stringEvent]()
		broker.maxSize = 0
		expected := "hello world"
		err := broker.Publish(ctx, stringEvent{
			ID:   uuid.New(),
			Data: expected,
		})
		require.ErrorIs(t, err, ErrSendFailed)
	})
}
