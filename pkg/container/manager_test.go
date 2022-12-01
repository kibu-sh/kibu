package container

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestManager(t *testing.T) {
	ctx := context.Background()
	manager, err := NewManager()
	require.NoError(t, err)
	defer manager.Cleanup(ctx)

	createParams := CreateParams{
		Name: "postgres",
		Container: &container.Config{
			Image: DefaultPostgresImage,
			Env: Environment(map[string]string{
				"POSTGRES_PASSWORD": "password",
			}).ToSlice(),
		},
		Host: &container.HostConfig{
			AutoRemove: true,
		},
	}

	container, err := manager.Create(ctx, createParams)
	require.NoError(t, err)
	require.NotNil(t, container)

	container2, err := manager.Create(ctx, createParams)
	require.NoError(t, err, "create should be idempotent")
	require.Equal(t, container.ID, container2.ID)

	_, err = manager.CreateAndStart(ctx, createParams)
	require.NoError(t, err)

	info, err := manager.client.ContainerInspect(ctx, container.ID)
	require.NoError(t, err)
	require.Equal(t, info.State.Status, "running")

	err = manager.client.ContainerKill(ctx, container.ID, "SIGKILL")
	require.NoError(t, err)

	_, err = manager.CreateAndStart(ctx, createParams)
	require.NoError(t, err)
}
