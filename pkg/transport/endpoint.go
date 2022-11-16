package transport

import "context"

// Endpoint is any function that can be modeled as service call.
// These should remain transport agnostic and are used to implement business logic.
type Endpoint interface {
	Serve(ctx context.Context, request any) (response any, err error)
	InitRequest() any
}

// EndpointFunc is a functional implementation of Endpoint
type EndpointFunc[Req, Res any] func(ctx context.Context, request Req) (response Res, err error)

// Serve implements Endpoint
func (e EndpointFunc[Req, Res]) Serve(ctx context.Context, request any) (response any, err error) {
	// pass by value allows the upstream endpoint to rely on non-nil values
	v := request.(*Req)
	return e(ctx, *v)
}

func (e EndpointFunc[Req, Res]) InitRequest() any {
	return new(Req)
}

// EndpointMiddlewareFunc decorates an Endpoint with additional functionality
type EndpointMiddlewareFunc func(next Endpoint) Endpoint

func NewEndpoint[Req, Res any](handler EndpointFunc[Req, Res]) Endpoint {
	return handler
}
