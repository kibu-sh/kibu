package storage

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

// compile check that CloudBucket implements Bucket
var _ Bucket = (*CloudBucket)(nil)

func TestBucket(t *testing.T) {
	tmpDir, dirErr := os.MkdirTemp("", "bucket-test")
	require.NoError(t, dirErr)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	tmpBucket := NewURL("file", tmpDir)

	t.Run("should create a file reference", func(t *testing.T) {
		bucket, err := NewCloudBucket(ctx, NewURL("gs", "my-bucket"))
		require.NoError(t, err)
		require.Equal(t, "gs://my-bucket/path/to/file", bucket.FileRef("path/to/file").Url().String())
	})

	t.Run("should create a write stream", func(t *testing.T) {
		bucket, err := NewCloudBucket(ctx, tmpBucket)
		require.NoError(t, err)

		stream, err := bucket.CreateWriter(context.Background(), bucket.FileRef("path/to/file"))
		require.NoError(t, err)

		_, err = stream.Write([]byte("hello world"))
		require.NoError(t, err)

		err = stream.Close()
		require.NoError(t, err)
		require.FileExists(t, filepath.Join(tmpDir, "/path/to/file"))
	})

	t.Run("should be able to create a reader", func(t *testing.T) {
		bucket, err := NewCloudBucket(ctx, tmpBucket)
		require.NoError(t, err)

		stream, err := bucket.CreateWriter(context.Background(), bucket.FileRef("path/to/file"))
		require.NoError(t, err)

		_, err = stream.Write([]byte("hello world"))
		require.NoError(t, err)

		err = stream.Close()
		require.NoError(t, err)
		require.FileExists(t, filepath.Join(tmpDir, "/path/to/file"))

		reader, err := bucket.CreateReader(context.Background(), bucket.FileRef("path/to/file"))
		require.NoError(t, err)

		_, err = reader.Read([]byte("hello world"))
		require.NoError(t, err)

		err = reader.Close()
		require.NoError(t, err)
	})

	t.Run("should be able to copy data from one bucket to another", func(t *testing.T) {
		srcBucket, err := NewCloudBucket(ctx, tmpBucket)
		require.NoError(t, err)

		srcFile := srcBucket.FileRef("path/to/file")
		srcWriter, err := srcBucket.CreateWriter(context.Background(), srcFile)
		require.NoError(t, err)

		_, err = srcWriter.Write([]byte("hello world"))
		require.NoError(t, err)

		err = srcWriter.Close()
		require.NoError(t, err)

		dstBucket, err := NewCloudBucket(ctx, tmpBucket)
		require.NoError(t, err)

		dstFile := dstBucket.FileRef("path/to/copy")
		err = Copy(context.Background(), srcFile, dstFile)
		require.NoError(t, err)
		require.FileExists(t, filepath.Join(tmpDir, "path/to/copy"))
	})
}

func TestParseURL(t *testing.T) {
	u, err := ParseURL("gs://my-bucket/path/to/directory")
	require.NoError(t, err)
	require.Equal(t, "gs", u.Driver)
	require.Equal(t, "my-bucket", u.Bucket)
	require.Equal(t, "/path/to/directory", u.Path)
}
