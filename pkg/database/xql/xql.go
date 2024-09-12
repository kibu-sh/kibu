package xql

import (
	"context"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/kibu-sh/kibu/pkg/ctxutil"
	"github.com/pkg/errors"
)

// builder func aliases

type Driver string

var (
	Postgres Driver = "postgres"
	SQLite3  Driver = "sqlite3"
	MySQL    Driver = "mysql"
)

type StatementBuilderType = sq.StatementBuilderType

var PostgresBuilder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
var DefaultBuilder = sq.StatementBuilder

func NewBuilder(driver Driver) (StatementBuilderType, error) {
	switch driver {
	case Postgres:
		return PostgresBuilder, nil
	case SQLite3, MySQL:
		return DefaultBuilder, nil
	default:
		return DefaultBuilder, errors.Errorf("%s is not a supported drievr", driver)
	}
}

var Alias = sq.Alias
var Case = sq.Case

// predicate aliases

type And = sq.And
type Or = sq.Or
type Eq = sq.Eq
type In = sq.Eq
type NotEq = sq.NotEq
type Gt = sq.Gt
type GtOrEq = sq.GtOrEq
type LtOrEq = sq.LtOrEq
type Lt = sq.Lt
type Like = sq.Like
type NotLike = sq.NotLike
type ILike = sq.ILike
type NotILike = sq.NotILike

// builder aliases

type Sqlizer = sq.Sqlizer
type SelectBuilder = sq.SelectBuilder
type InsertBuilder = sq.InsertBuilder
type UpdateBuilder = sq.UpdateBuilder
type DeleteBuilder = sq.DeleteBuilder

type QueryBuilder interface {
	SelectBuilder() SelectBuilder
	InsertBuilder() InsertBuilder
	UpdateBuilder() UpdateBuilder
	DeleteBuilder() DeleteBuilder
}

type EntityQueryBuilder[Entity any] interface {
	SelectOneBuilder(entity *Entity) SelectBuilder
	DeleteOneBuilder(entity *Entity) DeleteBuilder
	UpdateOneBuilder(entity *Entity) UpdateBuilder
}

type StatementBuilder interface {
	ToSql() (stm string, args []any, err error)
}
type SelectBuilderFunc func(q SelectBuilder) SelectBuilder
type InsertBuilderFunc func(q InsertBuilder) InsertBuilder
type UpdateBuilderFunc func(q UpdateBuilder) UpdateBuilder
type DeleteBuilderFunc func(q DeleteBuilder) DeleteBuilder

type QueryFunc func(ctx context.Context, dest any, q StatementBuilder)
type QueryStmFunc func(ctx context.Context, dest any, stm string, args ...any) error
type ExecFunc func(ctx context.Context, stm string, args ...any) (sql.Result, error)

func ExecAsQueryStmFunc(execFunc ExecFunc) QueryStmFunc {
	return func(ctx context.Context, dest any, stm string, args ...any) error {
		_, err := execFunc(ctx, stm, args...)
		return err
	}
}

type QueryWithParams struct {
	Target       any
	Query        StatementBuilder
	QueryStmFunc QueryStmFunc
}

func QueryWith(ctx context.Context, params QueryWithParams) (err error) {
	stm, args, err := params.Query.ToSql()
	if err != nil {
		return
	}
	return params.QueryStmFunc(ctx, params.Target, stm, args...)
}

type QueryRunner interface {
	GetContext(ctx context.Context, dest any, stm string, args ...any) error
	SelectContext(ctx context.Context, dest any, stm string, args ...any) error
}

type ExecRunner interface {
	ExecContext(ctx context.Context, stm string, args ...any) (result sql.Result, err error)
}

type Connection interface {
	QueryRunner
	ExecRunner
}

type RawSQLFunc func() (string, []any)

//revive:disable:var-naming
func (r RawSQLFunc) ToSql() (sql string, args []any, err error) {
	sql, args = r()
	return
}

//revive:enable:var-naming

func RawSQL(sql string, args ...any) RawSQLFunc {
	return func() (string, []any) {
		return sql, args
	}
}

func ApplySelectBuilderFuncs(s SelectBuilder, b ...SelectBuilderFunc) SelectBuilder {
	for _, f := range b {
		s = f(s)
	}

	return s
}

func ApplyInsertBuilderFuncs(s InsertBuilder, b ...InsertBuilderFunc) InsertBuilder {
	for _, f := range b {
		s = f(s)
	}

	return s
}

func ApplyUpdateBuilderFuncs(s UpdateBuilder, b ...UpdateBuilderFunc) UpdateBuilder {
	for _, f := range b {
		s = f(s)
	}

	return s
}

func ApplyDeleteBuilderFuncs(s DeleteBuilder, b ...DeleteBuilderFunc) DeleteBuilder {
	for _, f := range b {
		s = f(s)
	}

	return s
}

type connectionCtxKey struct{}

var ConnectionContextStore = ctxutil.NewStore[*sqlx.DB, connectionCtxKey]()
