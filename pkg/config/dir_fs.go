package config

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type DirectoryFS struct {
	Path string
}

func (d DirectoryFS) OpenReadable(ctx context.Context, params OpenParams) (stream io.ReadCloser, err error) {
	return os.Open(filepath.Join(d.Path, params.Path))
}

func (d DirectoryFS) OpenWritable(ctx context.Context, params OpenParams) (stream io.WriteCloser, err error) {
	return os.OpenFile(filepath.Join(d.Path, params.Path), os.O_CREATE|os.O_WRONLY, 0644)
}
