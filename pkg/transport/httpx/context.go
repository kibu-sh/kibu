package httpx

import (
	"context"
	"github.com/discernhq/devx/pkg/transport"
)

var _ transport.Context = (*Context)(nil)

type Context struct {
	context.Context
	req    *Request
	writer *Response
	codec  transport.Codec
}

func (c *Context) WithContext(ctx context.Context) transport.Context {
	c.Context = ctx
	return c
}

func (c *Context) Codec() transport.Codec {
	return c.codec
}

func (c *Context) Request() transport.Request {
	return c.req
}

func (c *Context) Response() transport.Response {
	return c.writer
}
