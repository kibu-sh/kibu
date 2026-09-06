package testingx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/kibu-sh/kibu/pkg/appcontext"
	"github.com/kibu-sh/kibu/pkg/container"
	"github.com/kibu-sh/kibu/pkg/ctxutil"
	"github.com/kibu-sh/kibu/pkg/netx"

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

func databasePath(database string) string {
	u := &url.URL{Path: database}
	return u.Path
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

	dsn = defaultDatabaseURL(address.HostPort(), PostgresUserinfo)

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
	resolved, joinErr := url.JoinPath("/", databasePath(params.Database))
	if joinErr != nil {
		err = joinErr
		return
	}
	dsn.Path = resolved

	return
}

func createReadWriteUser(ctx context.Context, db *sql.DB) (err error) {
	user := ReadWriteUserinfo.Username()
	password, _ := ReadWriteUserinfo.Password()

	checkUserQuery := fmt.Sprintf("select exists(select 1 from pg_roles where rolname='%s');", user)
	var exists bool
	err = db.QueryRowContext(ctx, checkUserQuery).Scan(&exists)
	if err != nil {
		return
	}

	if exists {
		return
	}

	recreateUserQuery := fmt.Sprintf(`
		drop user if exists %s;
		create user %s with password '%s';
		grant pg_write_all_data to %s;
		grant pg_read_all_data to %s;
	`, user, user, password, user, user)
	_, err = db.ExecContext(ctx, recreateUserQuery)
	if err != nil {
		return
	}

	return
}

func recreateTestDatabase(ctx context.Context, db *sql.DB, databaseName string) (err error) {
	dropQuery := fmt.Sprintf("drop database if exists %s;", databaseName)
	_, err = db.ExecContext(ctx, dropQuery)
	if err != nil {
		return
	}

	createQuery := fmt.Sprintf("create database %s;", databaseName)
	_, err = db.ExecContext(ctx, createQuery)
	if err != nil {
		return
	}

	return
}

var PostgresUserinfo = url.UserPassword("postgres", "password")
var ReadWriteUserinfo = url.UserPassword("read_write", "password")

func defaultDatabaseURL(hostPort string, userinfo *url.Userinfo) *url.URL {
	dsn := &url.URL{
		Path:   "/postgres", // initially connect to the default database
		Scheme: "postgres",
		Host:   hostPort,
		User:   userinfo,
	}
	query := dsn.Query()
	query.Set("sslmode", "disable")
	dsn.RawQuery = query.Encode()
	return dsn
}

type MigrationProvider func(dsn string) (*migrate.Migrate, error)

const defaultPostgresImage = "postgres:14"

func postgresImageOrDefault(image string) string {
	if image != "" {
		return image
	}
	return defaultPostgresImage
}

type SetupPostgresDatabaseConnectionParams struct {
	Manager        *container.Manager
	LoadMigrations MigrationProvider
	ContainerName  string
	DatabaseName   string
	ImageURL       string
}

func SetupPostgresDatabaseConnection(
	ctx context.Context,
	params SetupPostgresDatabaseConnectionParams,
) (dsn *url.URL, err error) {
	dsn, err = NewPostgresDB(ctx, params.Manager, NewPostgresDBParams{
		ImageURL:      postgresImageOrDefault(params.ImageURL),
		Database:      params.DatabaseName,
		ContainerName: params.ContainerName,
	})
	if err != nil {
		return
	}

	var migrations *migrate.Migrate
	var adminDNS = defaultDatabaseURL(dsn.Host, PostgresUserinfo)
	adminDNS.Path = dsn.Path
	migrations, err = params.LoadMigrations(adminDNS.String())
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
	ImageURL       string
}

func SetupTestMainWithDB(
	m *testing.M,
	params SetupTestMainWithDBParams,
) {
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
			ImageURL:       params.ImageURL,
		})
	CheckErrFatal(err)

	db, err := sqlx.ConnectContext(ctx, "postgres", dsn.String())
	CheckErrFatal(err)

	ctx = connectionContextStore.Save(ctx, Connection{
		DB:  db,
		URL: dsn,
	})
	appcontext.UpdateCache(ctx)

	m.Run()
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

// ExportConnection stores a database connection in the shared application
// context so that it can be retrieved later via GetDB.
func ExportConnection(ctx context.Context, conn Connection) context.Context {
	return connectionContextStore.Save(ctx, conn)
}
