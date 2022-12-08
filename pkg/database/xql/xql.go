package xql

import sq "github.com/Masterminds/squirrel"

// builder func aliases

var Select = sq.Select
var Insert = sq.Insert
var Update = sq.Update
var Delete = sq.Delete
var Alias = sq.Alias
var Case = sq.Case

// predicate aliases

type And = sq.And
type Or = sq.Or
type Eq = sq.Eq
type NotEq = sq.NotEq
type Gt = sq.Gt
type GtOrEq = sq.GtOrEq
type LtOrEq = sq.LtOrEq
type Lt = sq.Lt
type NotIn = sq.NotEq
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
