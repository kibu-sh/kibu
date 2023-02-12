package transport

import "context"

type Context interface {
	context.Context
	Codec() Codec
	Request() Request
	Response() Response
	WithContext(ctx context.Context) Context
}
