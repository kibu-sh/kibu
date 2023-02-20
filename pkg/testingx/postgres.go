package testingx

import (
	"context"
	"database/sql"
	"github.com/cenkalti/backoff"
	"github.com/discernhq/devx/pkg/appcontext"
	"github.com/discernhq/devx/pkg/container"
	"github.com/discernhq/devx/pkg/database/xql"
	"github.com/discernhq/devx/pkg/netx"
	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"testing"
	"time"

	containerapi "github.com/docker/docker/api/types/container"

	_ "github.com/lib/pq"
)

type NewPostgresDBParams struct {
	Database      string
	ImageURL      string
	ContainerName string
	Timeout       *time.Duration
}

func BackoffWithTimeout(
	ctx context.Context,
	timeout *time.Duration,
) backoff.BackOffContext {
	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = *timeout
	policy.Reset()
	return backoff.WithContext(policy, ctx)
}

func WaitForPostgres(ctx context.Context, db *sql.DB, timeout *time.Duration) container.StartOption {
	return func(container *container.Container) error {
		return backoff.Retry(func() error {
			return db.PingContext(ctx)
		}, BackoffWithTimeout(ctx, timeout))
	}
}

func NewPostgresDB(
	ctx context.Context,
	manager *container.Manager,
	params NewPostgresDBParams,
) (dsn *url.URL, err error) {
	free, err := netx.GetFreeAddr()
	if err != nil {
		return
	}

	if params.Timeout == nil {
		timeout := time.Second * 30
		params.Timeout = &timeout
	}

	pgContainer, err := manager.Create(ctx, container.CreateParams{
		Name: params.ContainerName,
		Container: &containerapi.Config{
			Image: params.ImageURL,
			Env: container.Environment(map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "password",
				"POSTGRES_DB":       params.Database,
			}).ToSlice(),
		},
		Host: &containerapi.HostConfig{
			PortBindings: nat.PortMap{
				"5432/tcp": []nat.PortBinding{
					{
						HostIP:   free.Host(),
						HostPort: free.Port(),
					},
				},
			},
		},
	})

	if err != nil {
		return
	}

	address, err := pgContainer.Address("5432/tcp")
	if err != nil {
		return
	}

	dsn = &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword("postgres", "password"),
		Host:   address.HostPort(),
		Path:   "/" + params.Database,
	}

	query := dsn.Query()
	query.Set("sslmode", "disable")

	dsn.RawQuery = query.Encode()

	db, err := sql.Open("postgres", dsn.String())
	if err != nil {
		return
	}

	defer func() {
		_ = db.Close()
	}()

	pgContainer, err = manager.Start(ctx, container.StartParams{
		Name:    params.ContainerName,
		Timeout: params.Timeout,
	}, WaitForPostgres(ctx, db, params.Timeout))

	return
}

type MigrationProvider func(dsn string) (*migrate.Migrate, error)

func SetupPostgresDatabaseConnection(
	ctx context.Context,
	manager *container.Manager,
	loadMigrations MigrationProvider,
	containerName string,
) (dsn *url.URL, err error) {
	dsn, err = NewPostgresDB(ctx, manager, NewPostgresDBParams{
		ImageURL:      "postgres:14",
		Database:      containerName,
		ContainerName: containerName,
	})
	if err != nil {
		return
	}

	var migrations *migrate.Migrate
	migrations, err = loadMigrations(dsn.String())
	if err != nil {
		return
	}

	err = migrations.Up()
	switch {
	case errors.Is(err, migrate.ErrNoChange):
		err = nil
		break
	case err != nil:
		return
	}

	return
}

func Context() context.Context {
	return appcontext.Context()
}

func SetupTestMainWithDB(
	m *testing.M,
	loadMigrations MigrationProvider,
	containerName string,
) {
	var code int
	ctx := Context()
	sharedManager, err := container.NewManager()
	CheckErrFatal(err)
	appcontext.UpdateCache(
		container.ManagerContextStore.Save(ctx, sharedManager),
	)

	dsn, err := SetupPostgresDatabaseConnection(
		ctx,
		sharedManager,
		loadMigrations,
		containerName,
	)
	CheckErrFatal(err)

	db, err := sqlx.ConnectContext(ctx, "postgres", dsn.String())
	CheckErrFatal(err)
	appcontext.UpdateCache(
		xql.ConnectionContextStore.Save(ctx, db),
	)

	code = m.Run()
	_ = sharedManager.Cleanup(ctx)

	os.Exit(code)
}
