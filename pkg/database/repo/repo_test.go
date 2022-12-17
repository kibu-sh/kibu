package repo

import (
	"context"
	"database/sql"
	"github.com/discernhq/devx/pkg/database"
	"github.com/discernhq/devx/pkg/database/repo/testdata/testmodels"
	. "github.com/discernhq/devx/pkg/database/xql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func copyDatabase() (dbCopyPath string, cleanup func(), err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", func() {}, err
	}

	testdata := filepath.Join(cwd, "testdata")
	srcDB := filepath.Join(testdata, "chinook.db")
	tmpDir, err := os.MkdirTemp("", "testdb")
	if err != nil {
		return "", func() {}, err
	}

	cleanup = func() { _ = os.RemoveAll(tmpDir) }
	dbCopyPath = filepath.Join(tmpDir, filepath.Base(srcDB))

	src, err := os.Open(srcDB)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.Create(dbCopyPath)
	if err != nil {
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return
}

func TestRepository(t *testing.T) {
	ctx := context.Background()
	dbPath, cleanup, dbCpErr := copyDatabase()
	require.NoError(t, dbCpErr)
	defer cleanup()

	conn, connErr := database.NewConn(ctx, database.Sqlite3, dbPath)
	require.NoError(t, connErr)

	t.Run("should be able to find one", func(t *testing.T) {
		repo, _ := NewRepo[testmodels.Album]()
		album, err := repo.Query(conn).FindOne(ctx, &testmodels.Album{
			AlbumID: 1,
		})

		require.NoError(t, err)
		require.NotNil(t, album)
		require.Equal(t, 1, album.AlbumID)
		require.Equal(t, "For Those About To Rock We Salute You", album.Title)
	})

	t.Run("should be able to find many", func(t *testing.T) {
		repo, _ := NewRepo[testmodels.Album]()
		albums, err := repo.Query(conn).FindMany(ctx, func(q SelectBuilder) SelectBuilder {
			return q.Where(Like{"Title": "%rock%"})
		})

		require.NoError(t, err)
		require.Equal(t, &testmodels.Album{
			AlbumID:  216,
			ArtistID: 142,
			Title:    "Hot Rocks, 1964-1971 (Disc 1)",
		}, albums[6])
	})

	t.Run("should be able to save one", func(t *testing.T) {
		repo, _ := NewRepo[testmodels.Album]()

		expected := &testmodels.Album{
			AlbumID:  500,
			ArtistID: 1,
			Title:    "I'm new here... how are you? (Small Talk)",
		}

		err := repo.Query(conn).CreateOne(ctx, expected)
		require.NoError(t, err)

		album, err := repo.Query(conn).FindOne(ctx, &testmodels.Album{AlbumID: 500})
		require.NoError(t, err)
		require.Equal(t, expected, album)
	})

	t.Run("should be able to intercept results", func(t *testing.T) {
		queryHookRun := false
		privacyErr := errors.New("privacy error")
		queryHook := WithQueryHook(func(ctx Context, result any) error {
			queryHookRun = true
			require.Equal(t, result, &testmodels.Album{}, "should intercept zero-value result")
			return nil
		})

		privacyHook := WithResultHook(func(ctx Context, result any) error {
			require.True(t, queryHookRun, "should run query hook first")
			require.Equal(t, &testmodels.Album{
				AlbumID:  1,
				ArtistID: 1,
				Title:    "For Those About To Rock We Salute You",
			}, result, "should intercept non-zero result")
			return privacyErr
		})

		repo, _ := NewRepo[testmodels.Album](
			privacyHook,
			queryHook,
		)

		_, err := repo.Query(conn).FindOne(ctx, &testmodels.Album{AlbumID: 1})
		require.ErrorAs(t, err, &privacyErr)
	})

	t.Run("should be able to create one", func(t *testing.T) {
		repo, _ := NewRepo[testmodels.Album]()

		expected := &testmodels.Album{
			AlbumID:  600,
			ArtistID: 1,
			Title:    "I'm new here... how are you? (Small Talk)",
		}

		err := repo.Query(conn).CreateOne(ctx, expected)
		require.NoError(t, err)

		album, err := repo.Query(conn).FindOne(ctx, &testmodels.Album{AlbumID: 600})
		require.NoError(t, err)
		require.Equal(t, expected, album)
	})

	t.Run("should be able to create many", func(t *testing.T) {
		repo, _ := NewRepo[testmodels.Album]()

		expected := []*testmodels.Album{
			{
				AlbumID:  501,
				ArtistID: 1,
				Title:    "I'm new here... how are you? (Small Talk)",
			},
			{
				AlbumID:  502,
				ArtistID: 1,
				Title:    "Avoids eye contact (the sequel)",
			},
		}

		err := repo.Query(conn).CreateMany(ctx, expected)
		require.NoError(t, err)

		albums, err := repo.Query(conn).FindMany(ctx, func(q SelectBuilder) SelectBuilder {
			return q.Where(In{"AlbumID": []int{501, 502}})
		})
		require.NoError(t, err)
		require.Equal(t, expected, albums)
	})

	t.Run("should be able to save one", func(t *testing.T) {
		repo, _ := NewRepo[testmodels.Album]()
		album, err := repo.Query(conn).FindOne(ctx, &testmodels.Album{AlbumID: 1})
		require.NoError(t, err)

		album.Title = "Updated Title"
		err = repo.Query(conn).SaveOne(ctx, album)
		require.NoError(t, err)

		updatedAlbum, err := repo.Query(conn).FindOne(ctx, &testmodels.Album{AlbumID: 1})
		require.NoError(t, err)
		require.Equal(t, album, updatedAlbum)
	})

	t.Run("should be able to save many", func(t *testing.T) {
		repo, _ := NewRepo[testmodels.Album]()
		albums, err := repo.Query(conn).FindMany(ctx, func(q SelectBuilder) SelectBuilder {
			return q.Where(In{"AlbumID": []int{3, 4}})
		})
		require.NoError(t, err)

		for _, album := range albums {
			album.Title = "Updated Title"
		}

		err = repo.Query(conn).SaveMany(ctx, albums)
		require.NoError(t, err)

		updatedAlbums, err := repo.Query(conn).FindMany(ctx, func(q SelectBuilder) SelectBuilder {
			return q.Where(In{"AlbumID": []int{3, 4}})
		})
		require.NoError(t, err)
		require.Equal(t, albums, updatedAlbums)
	})

	t.Run("should be able to update many", func(t *testing.T) {
		repo, _ := NewRepo[testmodels.Album]()
		err := repo.Query(conn).UpdateMany(ctx, func(q UpdateBuilder) UpdateBuilder {
			return q.Where(In{
				"AlbumID": []int{3, 4},
			}).Set("Title", "Updated Title")
		})
		require.NoError(t, err)

		updatedAlbums, err := repo.Query(conn).FindMany(ctx, func(q SelectBuilder) SelectBuilder {
			return q.Where(In{"AlbumID": []int{3, 4}})
		})
		require.NoError(t, err)
		require.Equal(t, "Updated Title", updatedAlbums[0].Title)
		require.Equal(t, "Updated Title", updatedAlbums[1].Title)
	})

	t.Run("should be able to delete one", func(t *testing.T) {
		repo, _ := NewRepo[testmodels.Album]()
		err := repo.Query(conn).DeleteOne(ctx, &testmodels.Album{AlbumID: 1})
		require.NoError(t, err)

		_, err = repo.Query(conn).FindOne(ctx, &testmodels.Album{AlbumID: 1})
		require.Error(t, err)
	})

	t.Run("should be able to delete many", func(t *testing.T) {
		repo, _ := NewRepo[testmodels.Album]()
		err := repo.Query(conn).DeleteMany(ctx, func(q DeleteBuilder) DeleteBuilder {
			return q.Where(In{"AlbumID": []int{3, 4}})
		})
		require.NoError(t, err)

		_, err = repo.Query(conn).FindOne(ctx, &testmodels.Album{AlbumID: 3})
		require.ErrorIs(t, err, sql.ErrNoRows)

		albums, err := repo.Query(conn).FindMany(ctx, func(q SelectBuilder) SelectBuilder {
			return q.Where(In{"AlbumID": []int{3, 4}})
		})
		require.NoError(t, err)
		require.Empty(t, albums)
	})

	t.Run("should be able to create a repo with options", func(t *testing.T) {
		repo, err := NewRepo[testmodels.Album](
			WithLogger(noOpLogger{}),
		)
		require.NotNil(t, repo)
		require.NoError(t, err)
	})

	t.Run("repo hook should be called on find one", func(t *testing.T) {
		var queryOp Operation
		var resultOp Operation
		repo, _ := NewRepo[testmodels.Album](
			WithQueryHook(func(ctx Context, result any) error {
				queryOp = ctx.Operation()
				return nil
			}),
			WithQueryHook(func(ctx Context, result any) error {
				resultOp = ctx.Operation()
				return nil
			}),
		)
		_, _ = repo.Query(conn).FindOne(ctx, &testmodels.Album{AlbumID: 1})
		require.Equal(t, OpFindOne, queryOp)
		require.Equal(t, OpFindOne, resultOp)
	})
	t.Run("should be able to create handle transactions", func(t *testing.T) {
		tx, err := conn.BeginTxx(ctx, nil)
		require.NoError(t, err)

		repo, _ := NewRepo[testmodels.Album]()
		err = repo.Query(tx).DeleteOne(ctx, &testmodels.Album{AlbumID: 1})
		require.NoError(t, err)

		err = tx.Rollback()
		require.NoError(t, err)
	})

	t.Run("should be zero value when query hook returns an error", func(t *testing.T) {
		repo, _ := NewRepo[testmodels.Album](
			WithQueryHook(func(ctx Context, result any) error {
				return errors.New("FAIL")
			}),
		)
		album, err := repo.Query(conn).FindOne(ctx, &testmodels.Album{AlbumID: 1})
		require.Error(t, err)
		require.Equal(t, album, &testmodels.Album{})
	})
}

func TestHookChain(t *testing.T) {
	ctx := context.Background()
	dbPath, cleanup, dbCpErr := copyDatabase()
	require.NoError(t, dbCpErr)
	defer cleanup()

	conn, connErr := database.NewConn(ctx, database.Sqlite3, dbPath)
	require.NoError(t, connErr)

	t.Run("assert decorated hook order", func(t *testing.T) {
		var tErr = errors.New("test error")
		var decorate = HookDecorator(func(ctx Context, result any) error {
			return tErr
		})

		var called bool
		var query = decorate(func(ctx Context, result any) error {
			called = true
			return nil
		})

		require.False(t, called)
		require.ErrorAs(t, query(&OpContext{
			Context:   context.Background(),
			operation: OpFindMany,
			query:     nil,
		}, ""), &tErr)
	})

	t.Run("check all operations", func(t *testing.T) {
		for i := 0; i < int(OpEnd); i++ {
			called := false
			operation := Operation(i)
			repo, _ := NewRepo[testmodels.Album](
				WithQueryHook(func(ctx Context, result any) error {
					called = true
					require.Equal(t, operation, ctx.Operation())
					return nil
				}))

			query := repo.Query(conn)
			switch operation {
			case OpCreateOne:
				_ = query.CreateOne(ctx, &testmodels.Album{})
			case OpFindOne:
				_, _ = query.FindOne(ctx, &testmodels.Album{
					AlbumID: 1,
				})
			case OpSaveOne:
				_ = query.SaveOne(ctx, &testmodels.Album{})
			case OpFindMany:
				_, _ = query.FindMany(ctx, func(q SelectBuilder) SelectBuilder { return q })
			case OpUpdateMany:
				_ = query.UpdateMany(ctx, func(q UpdateBuilder) UpdateBuilder { return q })
			case OpDeleteOne:
				_ = query.DeleteOne(ctx, &testmodels.Album{AlbumID: 1})
			case OpDeleteMany:
				_ = query.DeleteMany(ctx, func(q DeleteBuilder) DeleteBuilder { return q })
			default:
				t.Fatalf("unknown operation: %s", operation)
			}
			require.Equalf(t, true, called, "executed %s hook to be called", operation)
		}

	})

}
