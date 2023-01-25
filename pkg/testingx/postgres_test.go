package testingx

import (
	"context"
	"database/sql"
	"github.com/discernhq/devx/pkg/container"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewPostgresDB(t *testing.T) {
	ctx := context.Background()
	manager, mErr := container.NewManager()
	require.NoError(t, mErr)
	defer manager.Cleanup(ctx)

	t.Run("should create a new postgres database and return a connection", func(t *testing.T) {
		dsn, err := NewPostgresDB(ctx, manager, NewPostgresDBParams{
			Database:      "test",
			ImageURL:      "postgres:latest",
			ContainerName: "test-postgres",
		})
		require.NoError(t, err)

		db, err := sql.Open("postgres", dsn.String())
		require.NoError(t, err)

		err = db.PingContext(ctx)
		require.NoError(t, err)
	})
}
