package transport

import "context"

// Handler is a transport agnostic handler the request could be cast back
type Handler interface {
	Serve(ctx context.Context, request Request, writer ResponseWriter) error
}

// MiddlewareFunc decorates a handler with additional functionality
type MiddlewareFunc func(next Handler) Handler

// HandlerFunc is a functional alias to the Handler interface
type HandlerFunc func(ctx context.Context, request Request, writer ResponseWriter) error

// Serve implements Handler
func (h HandlerFunc) Serve(ctx context.Context, request Request, writer ResponseWriter) error {
	return h(ctx, request, writer)
}
