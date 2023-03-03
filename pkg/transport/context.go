package transport

import "github.com/discernhq/devx/pkg/ctxutil"

type Context interface {
	Codec() Codec
	Request() Request
	Response() Response
}

type contextKey struct{}

var ContextStore = ctxutil.NewStore[Context, contextKey]()
