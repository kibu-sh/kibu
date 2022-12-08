package database

import (
	"context"
	"database/sql"
	"github.com/discernhq/devx/pkg/database/xql"
)

type Conn interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Get(ctx context.Context, dest any, query xql.Query) error
	Select(ctx context.Context, dest any, query xql.Query) error
	Exec(ctx context.Context, query xql.Query) (result sql.Result, err error)
	BeginTxn(ctx context.Context, opts *sql.TxOptions) (txn Txn, err error)
}

type Txn interface {
	Commit() error
	Rollback() error
}
