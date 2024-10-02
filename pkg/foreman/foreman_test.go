package foreman

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestManager(t *testing.T) {
	ctx := context.Background()
	manager := NewManager(ctx)

	var stopped bool
	err := manager.Register(NewProcess("proc1", func(ctx context.Context, ready func()) error {
		time.Sleep(time.Second * 2)
		ready()
		for {
			select {
			case <-ctx.Done():
				stopped = true
				return nil
			}
		}
	}))
	require.NoError(t, err)

	go func() {
		manager.Shutdown()
	}()
	err = manager.Wait()
	require.NoError(t, err)
	require.True(t, stopped)
}
