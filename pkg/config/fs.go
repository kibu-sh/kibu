package config

import (
	"context"
	"io"
)

type OpenParams struct {
	Path string
}

type FS interface {
	OpenReadable(ctx context.Context, params OpenParams) (stream io.ReadCloser, err error)
	OpenWritable(ctx context.Context, params OpenParams) (stream io.WriteCloser, err error)
}
