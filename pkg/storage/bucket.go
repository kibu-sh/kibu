package storage

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/url"
	"path/filepath"

	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
)

// URL urls usually have a protocol/schema
// This concrete structure contains information about the storage system itself
// When interacting with cloud systems this can be in the form of gs://bucket/fil
// When interacting with local systems this can be in the form of file://bucket/file
type URL struct {
	Driver string
	Bucket string
	Path   string
}

// NewURL is a convenience function for creating an instance of a URL
func NewURL(driver, bucket string) URL {
	return URL{
		Driver: driver,
		Bucket: bucket,
	}
}

// String serializes the URL to a string
// Usually this is in the form of driver://bucket/path
func (u URL) String() string {
	return fmt.Sprintf("%s://%s", u.Driver, filepath.Join(u.Bucket, u.Path))
}

// ParseURL is a convenience function for parsing a URL string
func ParseURL(bucketURL string) (parsed URL, err error) {
	u, err := url.Parse(bucketURL)

	if err != nil {
		err = errors.Wrapf(err, "failed to parse bucket url: %s", bucketURL)
		return
	}

	parsed = URL{
		Driver: u.Scheme,
		Bucket: u.Host,
		Path:   u.Path,
	}
	return
}

// Bucket is an abstraction for a storage bucket
// Implementers of this interface could reside locally or over networked storage
// This could be any cloud service or your local machines file system
type Bucket interface {
	Url() URL
	FileRef(path string) FileRef
	CreateWriter(ctx context.Context, ref FileRef) (writer io.WriteCloser, err error)
	CreateReader(ctx context.Context, ref FileRef) (reader io.ReadCloser, err error)
}

// FileRef is an abstraction for a file reference
// This interface doesn't hold a handle to any specific file
// It acts as a pointer to a file in storage
// It is the buckets responsibility to open handles and exposing them using the std io interfaces
type FileRef interface {
	Url() URL
	Path() string
	Bucket() Bucket
}

// Copy mimics the Go io.Copy function
// It is possible to copy FileRefs between buckets
//
// TODO: implement a CopyWithin function that is optimized for copying within the same bucket
// Most cloud providers have APIs for this
// This will require extending the Bucket interface to support this
// https://gocloud.dev/blob supports this
func Copy(ctx context.Context, src, dst FileRef) (err error) {
	reader, err := src.Bucket().CreateReader(ctx, src)
	if err != nil {
		return
	}
	defer reader.Close()

	writer, err := dst.Bucket().CreateWriter(ctx, dst)
	if err != nil {
		return
	}
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	return
}
