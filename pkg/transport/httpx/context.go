package httpx

import (
	"context"
	"github.com/discernhq/devx/pkg/transport"
)

type Context struct {
	context.Context
	req    *Request
	writer *Response
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
