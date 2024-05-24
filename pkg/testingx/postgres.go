package testingx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/discernhq/devx/pkg/appcontext"
	"github.com/discernhq/devx/pkg/container"
	"github.com/discernhq/devx/pkg/ctxutil"
	"github.com/discernhq/devx/pkg/database/xql"
	"github.com/discernhq/devx/pkg/netx"
	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
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
	tctx context.Context,
	timeout *time.Duration,
) backoff.BackOffContext {
	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = *timeout
	policy.Reset()
	return backoff.WithContext(policy, tctx)
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

	dsn = defaultDatabaseURL(address.HostPort())

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

	// create the logical database that will used by the test
	// this allows us to reuse a single container
	err = recreateTestDatabase(ctx, db, params.Database)
	if err != nil {
		return
	}

	// update the dsn to use the test database
	dsn.Path = fmt.Sprintf("/%s", params.Database)

	return
}

func recreateTestDatabase(ctx context.Context, db *sql.DB, databaseName string) (err error) {
	_, err = db.ExecContext(ctx, fmt.Sprintf("drop database if exists %s;", databaseName))
	if err != nil {
		return
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("create database %s;", databaseName))
	if err != nil {
		return
	}
	return
}

func defaultDatabaseURL(hostPort string) *url.URL {
	dsn := &url.URL{
		Path:   "/postgres", // initially connect to the default database
		Scheme: "postgres",
		Host:   hostPort,
		User:   url.UserPassword("postgres", "password"),
	}
	query := dsn.Query()
	query.Set("sslmode", "disable")
	dsn.RawQuery = query.Encode()
	return dsn
}

type MigrationProvider func(dsn string) (*migrate.Migrate, error)

type SetupPostgresDatabaseConnectionParams struct {
	Manager        *container.Manager
	LoadMigrations MigrationProvider
	ContainerName  string
	DatabaseName   string
}

func SetupPostgresDatabaseConnection(
	ctx context.Context,
	params SetupPostgresDatabaseConnectionParams,
) (dsn *url.URL, err error) {
	dsn, err = NewPostgresDB(ctx, params.Manager, NewPostgresDBParams{
		ImageURL:      "postgres:14",
		Database:      params.DatabaseName,
		ContainerName: params.ContainerName,
	})
	if err != nil {
		return
	}

	var migrations *migrate.Migrate
	migrations, err = params.LoadMigrations(dsn.String())
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

type SetupTestMainWithDBParams struct {
	LoadMigrations MigrationProvider
	ContainerName  string
	DatabaseName   string
}

func SetupTestMainWithDB(
	m *testing.M,
	params SetupTestMainWithDBParams,
) {
	var code int
	ctx := Context()
	sharedManager, err := container.NewManager()
	CheckErrFatal(err)
	ctx = container.ManagerContextStore.Save(ctx, sharedManager)
	appcontext.UpdateCache(ctx)

	dsn, err := SetupPostgresDatabaseConnection(
		ctx, SetupPostgresDatabaseConnectionParams{
			Manager:        sharedManager,
			LoadMigrations: params.LoadMigrations,
			ContainerName:  params.ContainerName,
			DatabaseName:   params.DatabaseName,
		})
	CheckErrFatal(err)

	db, err := sqlx.ConnectContext(ctx, "postgres", dsn.String())
	CheckErrFatal(err)
	ctx = xql.ConnectionContextStore.Save(ctx, db)
	appcontext.UpdateCache(ctx)

	ctx = connectionContextStore.Save(ctx, Connection{
		DB:  db,
		URL: dsn,
	})
	appcontext.UpdateCache(ctx)

	code = m.Run()
	// we stopped doing this to leave the container running
	// each call should supply its own unique database name
	// this will spin up a single container and create a logical database for each test
	// the user can now introspect the database after the test has run
	// this also reduces resource utilization in large tests
	//_ = sharedManager.Cleanup(ctx)

	os.Exit(code)
}

type Connection struct {
	DB  *sqlx.DB
	URL *url.URL
}
type connectionCtxKey struct{}

var connectionContextStore = ctxutil.NewStore[Connection, connectionCtxKey]()

func GetDB() (Connection, error) {
	return connectionContextStore.Load(Context())
}
