package buffchan

import (
	"context"
	"log/slog"
	"time"
)

type WindowBufferChannelOptions[T any] struct {
	Window        time.Duration
	Input         <-chan T
	Debounce      func(T) bool
	MaxBufferSize int
	WriteBehavior func(chan []T, []T)
}

func defaultWriteBehavior[T any](output chan []T, data []T) {
	output <- data
}

// DropOverflowWriteBehavior is a function that writes data to an output channel.
// If the output channel is full, the function will drop the data and continue.
// It is used to handle write overflow in a buffered channel.
func DropOverflowWriteBehavior[T any](output chan []T, data []T) {
	select {
	case output <- data:
	default:
		slog.Default().Debug("channel is full dropping overflow", slog.Int("size", len(data)))
	}
}

func NewWindowBufferedChannel[T any](
	ctx context.Context,
	opts WindowBufferChannelOptions[T],
) chan []T {
	output := make(chan []T, opts.MaxBufferSize)
	ticker := time.NewTicker(opts.Window)
	started := make(chan struct{})

	if opts.WriteBehavior == nil {
		opts.WriteBehavior = defaultWriteBehavior[T]
	}

	go func() {
		close(started)
		var buffer []T
		defer ticker.Stop()
		defer close(output)

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				if len(buffer) == 0 {
					continue
				}
				opts.WriteBehavior(output, buffer)
				buffer = make([]T, 0)

			case data, ok := <-opts.Input:
				if !ok {
					return
				}

				if opts.Debounce != nil && opts.Debounce(data) {
					ticker.Reset(opts.Window)
				}

				buffer = append(buffer, data)
			}
		}
	}()

	<-started
	return output
}
