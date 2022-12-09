package database

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"testing"

	. "github.com/discernhq/devx/pkg/database/xql"
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

	conn, connErr := NewConnection(ctx, "sqlite3", dbPath)
	require.NoError(t, connErr)

	t.Run("should be able to find one", func(t *testing.T) {
		repo, _ := NewRepo[Album, int](conn)
		album, err := repo.FindOne(ctx, 1)

		require.NoError(t, err)
		require.NotNil(t, album)
		require.Equal(t, 1, album.AlbumID)
		require.Equal(t, "For Those About To Rock We Salute You", album.Title)
	})

	t.Run("should be able to find many", func(t *testing.T) {
		repo, _ := NewRepo[Album, int](conn)
		albums, err := repo.FindMany(ctx, func(q SelectBuilder) Query {
			return q.Where(Like{"Title": "%rock%"})
		})

		require.NoError(t, err)
		require.Equal(t, &Album{
			AlbumID:  216,
			ArtistID: 142,
			Title:    "Hot Rocks, 1964-1971 (Disc 1)",
		}, albums[6])
	})

	t.Run("should be able to save one", func(t *testing.T) {
		repo, _ := NewRepo[Album, int](conn)

		expected := &Album{
			AlbumID:  500,
			ArtistID: 1,
			Title:    "I'm new here... how are you? (Small Talk)",
		}

		err := repo.CreateOne(ctx, expected)
		require.NoError(t, err)

		album, err := repo.FindOne(ctx, 500)
		require.NoError(t, err)
		require.Equal(t, expected, album)
	})

	t.Run("should be able to intercept results", func(t *testing.T) {
		queryHookRun := false
		privacyErr := errors.New("privacy error")
		queryHook := WithQueryHook(func(ctx Context, result any) error {
			queryHookRun = true
			require.Equal(t, result, &Album{}, "should intercept zero-value result")
			return nil
		})

		privacyHook := WithPrivacyHook(func(ctx Context, result any) error {
			require.True(t, queryHookRun, "should run query hook first")
			require.Equal(t, &Album{
				AlbumID:  1,
				ArtistID: 1,
				Title:    "For Those About To Rock We Salute You",
			}, result, "should intercept non-zero result")
			return privacyErr
		})

		repo, _ := NewRepo[Album, int](conn,
			privacyHook,
			queryHook,
		)

		_, err := repo.FindOne(ctx, 1)
		require.ErrorAs(t, err, &privacyErr)
	})

	t.Run("should be able to create one", func(t *testing.T) {
		repo, _ := NewRepo[Album, int](conn)

		expected := &Album{
			AlbumID:  600,
			ArtistID: 1,
			Title:    "I'm new here... how are you? (Small Talk)",
		}

		err := repo.CreateOne(ctx, expected)
		require.NoError(t, err)

		album, err := repo.FindOne(ctx, 600)
		require.NoError(t, err)
		require.Equal(t, expected, album)
	})

	t.Run("should be able to create many", func(t *testing.T) {
		repo, _ := NewRepo[Album, int](conn)

		expected := []*Album{
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

		err := repo.CreateMany(ctx, expected)
		require.NoError(t, err)

		albums, err := repo.FindMany(ctx, func(q SelectBuilder) Query {
			return q.Where(In{"AlbumID": []int{501, 502}})
		})
		require.NoError(t, err)
		require.Equal(t, expected, albums)
	})

	t.Run("should be able to save one", func(t *testing.T) {
		repo, _ := NewRepo[Album, int](conn)
		album, err := repo.FindOne(ctx, 1)
		require.NoError(t, err)

		album.Title = "Updated Title"
		err = repo.SaveOne(ctx, album)
		require.NoError(t, err)

		updatedAlbum, err := repo.FindOne(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, album, updatedAlbum)
	})

	t.Run("should be able to save many", func(t *testing.T) {
		repo, _ := NewRepo[Album, int](conn)
		albums, err := repo.FindMany(ctx, func(q SelectBuilder) Query {
			return q.Where(In{"AlbumID": []int{3, 4}})
		})
		require.NoError(t, err)

		for _, album := range albums {
			album.Title = "Updated Title"
		}

		err = repo.SaveMany(ctx, albums)
		require.NoError(t, err)

		updatedAlbums, err := repo.FindMany(ctx, func(q SelectBuilder) Query {
			return q.Where(In{"AlbumID": []int{3, 4}})
		})
		require.NoError(t, err)
		require.Equal(t, albums, updatedAlbums)
	})

	t.Run("should be able to update one", func(t *testing.T) {
		repo, _ := NewRepo[Album, int](conn)
		err := repo.UpdateOne(ctx, 1, func(q UpdateBuilder) Query {
			return q.Set("Title", "Updated Title")
		})
		require.NoError(t, err)

		updatedAlbum, err := repo.FindOne(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, "Updated Title", updatedAlbum.Title)
	})

	t.Run("should be able to update many", func(t *testing.T) {
		repo, _ := NewRepo[Album, int](conn)
		err := repo.UpdateMany(ctx, func(q UpdateBuilder) Query {
			return q.Where(In{"AlbumID": []int{3, 4}}).Set("Title", "Updated Title")
		})
		require.NoError(t, err)

		updatedAlbums, err := repo.FindMany(ctx, func(q SelectBuilder) Query {
			return q.Where(In{"AlbumID": []int{3, 4}})
		})
		require.NoError(t, err)
		require.Equal(t, "Updated Title", updatedAlbums[0].Title)
		require.Equal(t, "Updated Title", updatedAlbums[1].Title)
	})

	t.Run("should be able to delete one", func(t *testing.T) {
		repo, _ := NewRepo[Album, int](conn)
		err := repo.DeleteOne(ctx, 1)
		require.NoError(t, err)

		_, err = repo.FindOne(ctx, 1)
		require.Error(t, err)
	})

	t.Run("should be able to delete many", func(t *testing.T) {
		repo, _ := NewRepo[Album, int](conn)
		err := repo.DeleteMany(ctx, func(q DeleteBuilder) Query {
			return q.Where(In{"AlbumID": []int{3, 4}})
		})
		require.NoError(t, err)

		albums, err := repo.FindMany(ctx, func(q SelectBuilder) Query {
			return q.Where(In{"AlbumID": []int{3, 4}})
		})
		require.NoError(t, err)
		require.Empty(t, albums)
	})

	t.Run("should be able to count", func(t *testing.T) {
		// repo, _ := NewRepo[Album, int](conn)
		// count, err := repo.Count(ctx, func(q SelectBuilder) Query {
		// 	return q.Where(In{"AlbumID": []int{3, 4}})
		// })
		// require.NoError(t, err)
		// require.Equal(t, 2, count)
	})

	t.Run("check all operations", func(t *testing.T) {
		for i := 0; i < int(OpEnd); i++ {
			operation := Operation(i)
			repo, _ := NewRepo[Album, int](conn,
				WithQueryHook(func(ctx Context, result any) error {
					require.Equal(t, operation, ctx.Operation())
					return nil
				}))

			switch operation {
			case OpCreateOne:
				_ = repo.CreateOne(ctx, &Album{})
			case OpCreateMany:
				_ = repo.CreateMany(ctx, []*Album{})
			case OpFindOne:
				_, _ = repo.FindOne(ctx, 1)
			case OpFindMany:
				_, _ = repo.FindMany(ctx, func(q SelectBuilder) Query { return q })
			case OpSaveOne:
				_ = repo.SaveOne(ctx, &Album{})
			case OpSaveMany:
				_ = repo.SaveMany(ctx, []*Album{})
			case OpUpdateOne:
				_ = repo.UpdateOne(ctx, 1, func(q UpdateBuilder) Query { return q })
			case OpUpdateMany:
				_ = repo.UpdateMany(ctx, func(q UpdateBuilder) Query { return q })
			case OpDeleteOne:
				_ = repo.DeleteOne(ctx, 1)
			case OpDeleteMany:
				_ = repo.DeleteMany(ctx, func(q DeleteBuilder) Query { return q })
			default:
				t.Fatalf("unknown operation: %s", operation)
			}
		}

	})

	t.Run("should be able to find one to one relation", func(t *testing.T) {})
	t.Run("should be able to find one to many relation", func(t *testing.T) {})
	t.Run("should be able to find many to many relation", func(t *testing.T) {})
	t.Run("should be able to find many to one relation", func(t *testing.T) {})
	t.Run("should be able to find many to one relation", func(t *testing.T) {})
	t.Run("should be able to find one through relation", func(t *testing.T) {})
	t.Run("should be able to find many through relation", func(t *testing.T) {})

	t.Run("should be able to create a repo with options", func(t *testing.T) {
		repo, err := NewRepo[Album, int](conn,
			WithLogger(noOpLogger{}),
		)
		require.NotNil(t, repo)
		require.NoError(t, err)
	})

	t.Run("repo hook should be called on find one", func(t *testing.T) {
		var queryOp Operation
		var resultOp Operation
		repo, _ := NewRepo[Album, int](conn,
			WithQueryHook(func(ctx Context, result any) error {
				queryOp = ctx.Operation()
				return nil
			}),
			WithQueryHook(func(ctx Context, result any) error {
				resultOp = ctx.Operation()
				return nil
			}),
		)
		_, _ = repo.FindOne(ctx, 1)
		require.IsType(t, OpFindOne, queryOp)
		require.IsType(t, OpFindOne, resultOp)
	})

	t.Run("model should be zero value when query hook returns an error", func(t *testing.T) {
		repo, _ := NewRepo[Album, int](conn,
			WithQueryHook(func(ctx Context, result any) error {
				return errors.New("FAIL")
			}),
		)
		album, err := repo.FindOne(ctx, 1)
		require.Error(t, err)
		require.Equal(t, album, &Album{})
	})
}

func TestHookChain(t *testing.T) {
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
}
