package transport

import "context"

type RawEndpoint struct {
	HandlerFunc HandlerFunc
	Middleware  []Middleware
}

func NewRawEndpoint(
	endpointFunc HandlerFunc,
) *RawEndpoint {
	return &RawEndpoint{
		HandlerFunc: endpointFunc,
	}
}

func (endpoint *RawEndpoint) WithMiddleware(middleware ...Middleware) *RawEndpoint {
	endpoint.Middleware = middleware
	return endpoint
}

func newRawEndpointFunc(endpointFunc HandlerFunc) EndpointFunc[any, any] {
	return func(ctx context.Context, request any) (response any, err error) {
		tCtx, err := ContextStore.Load(ctx)
		if err != nil {
			return
		}
		if err = endpointFunc(tCtx); err == nil {
			err = ErrResponseIntercepted
		}
		return
	}
}

func (endpoint *RawEndpoint) Serve(tctx Context) (err error) {
	err = ApplyMiddleware(endpoint.HandlerFunc, endpoint.Middleware...).Serve(tctx)
	return
}
