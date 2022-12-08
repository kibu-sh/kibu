package storage

import (
	"context"
	"gocloud.dev/blob"
	"io"
)

// NewCloudBucket opens a bucket (this could be in the cloud or local)
// This operation will return an error if the driver in the provided URL isn't supported
// This can be one of gs, file, s3, etc...
// See https://gocloud.dev/blob for more information
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

// CloudBucket is a concrete implementation of the Bucket interface
// It adapts the gocloud.dev/blob.Bucket interface to our internal Bucket interface
type CloudBucket struct {
	driver string
	bucket string
	store  *blob.Bucket
}

// Url only returns the base URL for the bucket itself not a FileRef
func (c *CloudBucket) Url() URL {
	return URL{
		Driver: c.driver,
		Bucket: c.bucket,
	}
}

// FileRef takes a path and returns a FileRef
// This should be relative to the buckets root
func (c *CloudBucket) FileRef(path string) FileRef {
	return &CloudFileRef{
		path:   path,
		bucket: c,
	}
}

// CreateWriter takes a FileRef and returns a writer where content can be written
// It is the callers responsibility to close the writer when they're done
// Failing to do so can result in incomplete writes
// In the case of google cloud storage the file upload is transactional it will fail silently
func (c *CloudBucket) CreateWriter(ctx context.Context, ref FileRef) (writer io.WriteCloser, err error) {
	return c.store.NewWriter(ctx, ref.Path(), nil)
}

// CreateReader takes a FileRef and returns a reader where content can be read
// It is the callers responsibility to close the reader when they're done
func (c *CloudBucket) CreateReader(ctx context.Context, ref FileRef) (reader io.ReadCloser, err error) {
	return c.store.NewReader(ctx, ref.Path(), nil)
}
