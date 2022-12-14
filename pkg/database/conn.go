package database

import (
	"context"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Driver string

var (
	Sqlite3  Driver = "sqlite3"
	Postgres Driver = "postgres"
)

func NewConn(ctx context.Context, driver Driver, dsn string) (*sqlx.DB, error) {
	return sqlx.ConnectContext(ctx, string(driver), dsn)
}
