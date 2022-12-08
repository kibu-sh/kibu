package database

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"

	sq "github.com/Masterminds/squirrel"

	_ "github.com/mattn/go-sqlite3"
)

type Artist struct {
	ArtistID int    `db:"ArtistId"`
	Name     string `db:"Name"`
	// Albums   []Album `db:",rel=albums,ref=ArtistId"`
}

type Album struct {
	AlbumID  int    `db:"AlbumId,pk,table=albums"`
	ArtistID int    `db:"ArtistId"`
	Title    string `db:"Title"`
	Omitted  string `db:"-"`
	// Artist   Artist `db:",rel=artist,fields=[ArtistId],ref=[ArtistId]"`
}

var AlbumDefinition = EntityDefinition{
	Table: "albums",
}

func TestQuery(t *testing.T) {
	cwd, _ := os.Getwd()
	ctx := context.Background()
	testdata := filepath.Join(cwd, "testdata")
	conn, connErr := sqlx.Open("sqlite3", filepath.Join(testdata, "chinook.conn"))
	require.NoError(t, connErr)

	t.Run("should be able to bind entity to query results", func(t *testing.T) {
		album, err := ExecSQL[Album](ctx, ExecWith(conn.GetContext), RawSQL(
			"select * from albums where AlbumId = ?", 1,
		))
		require.NoError(t, err)
		require.Equal(t, 1, album.AlbumID)
		require.Contains(t, album.Title, "About To Rock")
	})

	t.Run("should be able to find entity using fluent builder", func(t *testing.T) {
		query := sq.Select("*").From("albums").Where("AlbumId = ?", 1)
		album, err := ExecSQL[Album](ctx, ExecWith(conn.GetContext), query)
		require.NoError(t, err)
		require.Equal(t, 1, album.AlbumID)
		require.Contains(t, album.Title, "About To Rock")
	})

	t.Run("should find one by primary key", func(t *testing.T) {
		t.Skip()
		findOne := NewSQLFind[Album, int](conn)
		album, err := findOne(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, 1, album.AlbumID)
		require.Contains(t, album.Title, "About To Rock")
	})

	t.Run("should return ErrNotFound find one or throw", func(t *testing.T) {
		t.Skip()
		findOne := NewSQLFind[Album, int](conn)
		_, err := findOne(ctx, 1000)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("should be be able to handle transactions", func(t *testing.T) {

	})

	t.Run("should find one by where query", func(t *testing.T) {})
	t.Run("should find many by where query", func(t *testing.T) {})
	t.Run("should be able to intercept result", func(t *testing.T) {})
	t.Run("should be able to query related entities", func(t *testing.T) {})
	// t.Run("should be able to intercept query", func(t *testing.T) {
	// 	findOne := NewSQLFind[Album, int](conn)
	// 	findOne = applyFindOne[Album, int](findOne, NewFindFuncMiddleware(func(ctx context.Context, key int) (*Album, error) {
	// 		return nil, errors.Errorf("intercepted")
	// 	}))
	// 	album, err := findOne(ctx, 1)
	// 	require.Error(t, err)
	// 	require.Nil(t, album)
	// })
}
