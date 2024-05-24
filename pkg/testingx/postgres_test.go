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
	//defer manager.Cleanup(ctx)

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

		t.Run("should create a second postgres database in the same container", func(t *testing.T) {
			dsn2, err := NewPostgresDB(ctx, manager, NewPostgresDBParams{
				Database:      "test2",
				ImageURL:      "postgres:latest",
				ContainerName: "test-postgres",
			})
			require.NoError(t, err)
			require.Equal(t, dsn.Host, dsn2.Host)
			require.Equal(t, dsn.User, dsn2.User)
			require.NotEqualf(t, dsn.Path, dsn2.Path, "expected different database names")

			db, err := sql.Open("postgres", dsn2.String())
			require.NoError(t, err)

			err = db.PingContext(ctx)
			require.NoError(t, err)
		})
	})

	// write a unit test that checks there is only a single container with the name "test-postgres"
	// this should also attempt to create another database with a different name and check that each DB exists

}
