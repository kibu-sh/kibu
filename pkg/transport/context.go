package transport

import "github.com/kibu-sh/kibu/pkg/ctxutil"

type Context interface {
	Codec() Codec
	Request() Request
	Response() Response
}

type contextKey struct{}

var ContextStore = ctxutil.NewStore[Context, contextKey]()
