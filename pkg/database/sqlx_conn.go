package database

import (
	"context"
	"database/sql"
	"github.com/discernhq/devx/pkg/database/xql"
	"github.com/jmoiron/sqlx"
)

var _ Conn = (*SQLXConn)(nil)

type SQLXConn struct {
	db *sqlx.DB
}

func (s *SQLXConn) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *SQLXConn) Close(ctx context.Context) error {
	return s.db.Close()
}

func (s *SQLXConn) Get(ctx context.Context, dest any, query xql.Query) (err error) {
	stm, args, err := query.ToSql()
	if err != nil {
		return
	}
	return s.db.GetContext(ctx, dest, stm, args...)
}

func (s *SQLXConn) Select(ctx context.Context, dest any, query xql.Query) error {
	stm, args, err := query.ToSql()
	if err != nil {
		return err
	}
	return s.db.SelectContext(ctx, dest, stm, args...)
}

func (s *SQLXConn) BeginTxn(ctx context.Context, opts *sql.TxOptions) (txn Txn, err error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	txn = &SQLXTxn{tx}
	return
}

func (s *SQLXConn) Exec(ctx context.Context, query xql.Query) (result sql.Result, err error) {
	stm, args, err := query.ToSql()
	if err != nil {
		return
	}
	return s.db.ExecContext(ctx, stm, args...)
}

type SQLXTxn struct {
	*sqlx.Tx
}

func (s *SQLXTxn) Commit() error {
	return s.Tx.Commit()
}

func (s *SQLXTxn) Rollback() error {
	return s.Tx.Rollback()
}

func NewConnection(ctx context.Context, driver, dsn string) (conn *SQLXConn, err error) {
	db, err := sqlx.ConnectContext(ctx, driver, dsn)
	conn = &SQLXConn{db}
	return
}
