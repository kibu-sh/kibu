package storage

import (
	"context"
	"fmt"
	"gocloud.dev/blob"
	"io"
	"path/filepath"

	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
)

type URL struct {
	Driver string
	Bucket string
	Path   string
}

func NewURL(driver, bucket string) URL {
	return URL{
		Driver: driver,
		Bucket: bucket,
	}
}

func (u URL) String() string {
	return fmt.Sprintf("%s://%s", u.Driver, filepath.Join(u.Bucket, u.Path))
}

type Bucket interface {
	Url() URL
	FileRef(path string) FileRef
	CreateWriter(ctx context.Context, ref FileRef) (writer io.WriteCloser, err error)
	CreateReader(ctx context.Context, ref FileRef) (reader io.ReadCloser, err error)
}

type FileRef interface {
	Url() URL
	Path() string
}

func NewCloudBucket(ctx context.Context, url URL) (bucket *CloudBucket, err error) {
	store, err := blob.OpenBucket(ctx, url.String())
	if err != nil {
		return
	}
	return &CloudBucket{
		store:  store,
		bucket: url.Bucket,
		driver: url.Driver,
	}, nil
}

type CloudBucket struct {
	driver string
	bucket string
	store  *blob.Bucket
}

func (c *CloudBucket) Url() URL {
	return URL{
		Driver: c.driver,
		Bucket: c.bucket,
	}
}

type CloudFileRef struct {
	path   string
	bucket *CloudBucket
}

func (c *CloudFileRef) Path() string {
	return c.path
}

func (c *CloudFileRef) Url() URL {
	return URL{
		Driver: c.bucket.driver,
		Bucket: c.bucket.bucket,
		Path:   c.path,
	}
}

func (c *CloudBucket) FileRef(path string) FileRef {
	return &CloudFileRef{
		path:   path,
		bucket: c,
	}
}

func (c *CloudBucket) CreateWriter(ctx context.Context, ref FileRef) (writer io.WriteCloser, err error) {
	return c.store.NewWriter(ctx, ref.Path(), nil)
}

func (c *CloudBucket) CreateReader(ctx context.Context, ref FileRef) (reader io.ReadCloser, err error) {
	return c.store.NewReader(ctx, ref.Path(), nil)
}
