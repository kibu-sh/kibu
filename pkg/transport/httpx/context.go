package httpx

import (
	"github.com/discernhq/devx/pkg/transport"
)

var _ transport.Context = (*Context)(nil)

type Context struct {
	req    *Request
	writer *ResponseWriter
	codec  transport.Codec
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
