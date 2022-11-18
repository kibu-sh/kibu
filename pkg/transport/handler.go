package transport

// Handler is a transport agnostic handler the request could be cast back
type Handler interface {
	Serve(ctx Context) (err error)
}

// HandlerFunc is a functional alias to the Handler interface
type HandlerFunc func(ctx Context) (err error)

// Serve implements Handler
func (h HandlerFunc) Serve(ctx Context) error {
	return h(ctx)
}

type Middleware func(Handler) Handler

func ApplyMiddleware(handler Handler, middleware ...Middleware) Handler {
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

// NewMiddleware is a convenience method for implementing the function middleware pattern
// Provide a simple HandlerFunc if it doesn't return an error during a request the next Handler will be called
func NewMiddleware(handler HandlerFunc) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx Context) error {
			if err := handler(ctx); err != nil {
				return err
			}
			return next.Serve(ctx)
		})
	}
}
