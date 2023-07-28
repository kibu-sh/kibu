package config

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

var _ FS = (*DirectoryFS)(nil)

type DirectoryFS struct {
	Path string
}

func (d DirectoryFS) Root() string {
	return d.Path
}

func (d DirectoryFS) OpenReadable(ctx context.Context, params OpenParams) (stream io.ReadCloser, err error) {
	return os.Open(filepath.Join(d.Path, params.Path))
}

func (d DirectoryFS) OpenWritable(ctx context.Context, params OpenParams) (stream io.WriteCloser, err error) {
	file := filepath.Join(d.Path, params.Path)
	if err = os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return
	}
	return os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
}
