package database

import (
	"context"
	"github.com/Masterminds/squirrel"
	entity2 "github.com/discernhq/devx/pkg/database/entity"
	"github.com/jmoiron/sqlx"
)

type Query2 interface {
	String() string
}

type FindFunc[E, K any] func(ctx context.Context, key K) (entity *E, err error)
type FindAllFunc[E any] func(query Query2) (entities []*E, err error)

// type FindFuncMiddleware[E, K any] func(next FindFunc[E, K]) FindFunc[E, K]
//
// func NewFindFuncMiddleware[E, K any](middlewares ...FindFuncMiddleware[E, K]) FindFuncMiddleware[E, K] {
// 	return func(next FindFunc[E, K]) FindFunc[E, K] {
// 		for i := len(middlewares) - 1; i >= 0; i-- {
// 			next = middlewares[i](next)
// 		}
// 		return next
// 	}
// }
//
// func NewFindFunc[E, K any](db *sqlx.DB, table string, key string, middlewares ...FindFuncMiddleware[E, K]) FindFunc[E, K] {
// 	return NewFindFuncMiddleware[E, K](middlewares...)(NewSQLFind(db, table, key))
// }
//
// func NewSQLFind[E, K any](db *sqlx.DB, table string, key string) FindFunc[E, K] {
// 	return func(ctx context.Context, key K) (entity *E, err error) {
// 		entity = &E{}
// 		err = db.GetContext(ctx, entity, sq.Select("*").From(table).Where(sq.Eq{key: key}).ToSql())
// 		return
// 	}
// }
//
// func NewFindFunc[E, K any](db *sqlx.DB, table string, key string, middlewares ...FindFuncMiddleware[E, K]) FindFunc[E, K] {
// 	return NewFindFuncMiddleware[E, K](middlewares...)(NewSQLFind(db, table, key))
// }
//
// func NewSQLFind[E, K any](db *sqlx.DB, table string, key string) FindFunc[E, K] {
// 	return func(ctx context.Context, key K) (entity *E, err error) {
// 		entity = &E{}
// 		err = db.GetContext(ctx, entity, sq.Select("*").From(table).Where(sq.Eq{key: key}).ToSql())
// 		return
// 	}
// }
//
// func NewFindFunc[E, K any](db *sqlx.DB, table string, key string, middlewares ...FindFuncMiddleware[E, K]) FindFunc[E, K] {
// 	return NewFindFuncMiddleware[E, K](middlewares...)(NewSQLFind(db, table, key))
// }
//
// func NewSQLFind[E, K
// func NewFindFuncMiddleware[E Entity, K any](base FindFunc[E, K]) FindFuncMiddleware[E, K] {
// 	return func(next FindFunc[E, K]) FindFunc[E, K] {
// 		return func(ctx context.Context, key K) (entity *E, err error) {
// 			if entity, err = base(ctx, key); err != nil {
// 				return
// 			}
// 			return next(ctx, key)
// 		}
// 	}
// }
//
// func applyFindOne[E Entity, K any](base FindFunc[E, K], middleware ...FindFuncMiddleware[E, K]) (result FindFunc[E, K]) {
// 	result = base
// 	for _, m := range middleware {
// 		result = m(result)
// 	}
// 	return
// }

func NewSQLFind[E, K any](db *sqlx.DB) FindFunc[E, K] {
	return func(ctx context.Context, key K) (entity *E, err error) {
		entity = new(E)

		def, err := entity2.ReflectEntityDefinition[E, K]("db")
		if err != nil {
			return nil, err
		}

		sql, args, err := squirrel.Select(def.Fields().Names()...).
			From("table").
			Where(squirrel.Eq{"AlbumId": key}).
			ToSql()

		if err != nil {
			return nil, err
		}

		err = db.Get(entity, sql, args...)
		return
	}
}

type ExecFunc[E any] func(ctx context.Context) (entity *E, err error)

type SQLQuery interface {
	ToSql() (sql string, args []any, err error)
}

type RawSQLFunc func() (string, []any)

//revive:disable:var-naming
func (r RawSQLFunc) ToSql() (sql string, args []any, err error) {
	sql, args = r()
	return
}

//revive:enable:var-naming

type SQLExec interface {
	Exec(ctx context.Context, target any, query SQLQuery) error
}

type SQLExecFunc func(ctx context.Context, target any, query SQLQuery) error

func (f SQLExecFunc) Exec(ctx context.Context, target any, query SQLQuery) error {
	return f(ctx, target, query)
}

type BindFunc func(ctx context.Context, target any, sql string, args ...any) error

func ExecWith(bind BindFunc) SQLExecFunc {
	return func(ctx context.Context, target any, query SQLQuery) (err error) {
		stm, args, err := query.ToSql()
		if err != nil {
			return
		}
		err = bind(ctx, target, stm, args...)
		return
	}
}

func RawSQL(sql string, args ...any) RawSQLFunc {
	return func() (string, []any) {
		return sql, args
	}
}

func ExecSQL[E any](ctx context.Context, db SQLExec, query SQLQuery) (entity *E, err error) {
	entity = new(E)
	err = db.Exec(ctx, entity, query)
	return
}
