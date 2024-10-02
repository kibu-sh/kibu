package transport

// Handler is a transport agnostic handler the request could be cast back
type Handler interface {
	Serve(tctx Context) (err error)
}

// HandlerFunc is a functional alias to the Handler interface
type HandlerFunc func(tctx Context) (err error)

// Serve implements Handler
func (h HandlerFunc) Serve(tctx Context) error {
	return h(tctx)
}

type Middleware func(Handler) Handler
type MiddlewareFunc func(tctx Context, next Handler) error

func ApplyMiddleware(handler Handler, middleware ...Middleware) Handler {
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

// NewMiddleware is a convenience method for implementing the function middleware pattern
// Provide a simple HandlerFunc if it doesn't return an error during a request the next Handler will be called
func NewMiddleware(middlewareFunc MiddlewareFunc) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(tctx Context) error {
			return middlewareFunc(tctx, next)
		})
	}
}
