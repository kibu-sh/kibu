package buffchan

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewWindowBufferedChannel(t *testing.T) {
	tests := map[string]struct {
		window   time.Duration
		input    []int
		expected []int
		debounce func(int) bool
		timeout  time.Duration
	}{
		"with time window": {
			window:   2 * time.Second,
			timeout:  10 * time.Second,
			input:    []int{1, 2, 3},
			expected: []int{1, 2, 3},
			debounce: func(_ int) bool { return false },
		},
		"with debounce": {
			window:   2 * time.Second,
			timeout:  10 * time.Second,
			input:    []int{1, 2, 3},
			expected: []int{1, 2, 3},
			debounce: func(val int) bool { return val%2 == 0 },
		},
		"should be nil because debounce window exceeds timeout": {
			window:  2 * time.Second,
			timeout: 1 * time.Second,
			input:   []int{1, 2, 3},
			// debounce resets the window 3 times, so the output will be nil
			debounce: func(_ int) bool {
				return true
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			inputChan := make(chan int)
			outputChan := NewWindowBufferedChannel(ctx, WindowBufferChannelOptions[int]{
				Window:   tc.window,
				Input:    inputChan,
				Debounce: tc.debounce,
			})

			go func() {
				for _, val := range tc.input {
					inputChan <- val
				}
			}()

			require.Equal(t, tc.expected, <-outputChan)
		})
	}
}
