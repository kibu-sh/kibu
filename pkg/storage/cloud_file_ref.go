package storage

// CloudFileRef is a concrete implementation of the FileRef interface
// It adapts the gocloud.dev/blob.Bucket interface to our internal FileRef interface
type CloudFileRef struct {
	path   string
	bucket *CloudBucket
}

// Bucket returns a pointer to the bucket that this file reference belongs to
func (c *CloudFileRef) Bucket() Bucket {
	return c.bucket
}

// Path returns the path of the file reference relative to the bucket
// Given a URL of gs://bucket/path/to/file the path would be path/to/file
func (c *CloudFileRef) Path() string {
	return c.path
}

// Url returns the URL for the file reference including the bucket as the base
// gs://bucket/path/to/file
func (c *CloudFileRef) Url() URL {
	return URL{
		Driver: c.bucket.driver,
		Bucket: c.bucket.bucket,
		Path:   c.path,
	}
}
