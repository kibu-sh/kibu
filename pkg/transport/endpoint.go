package transport

import "context"

// EndpointFunc is a functional implementation of Endpoint
type EndpointFunc[Req, Res any] func(ctx context.Context, request Req) (response Res, err error)

// Endpoint is any function that can be modeled as service call.
// These should remain transport agnostic and are used to implement business logic.
type Endpoint[Req, Res any] struct {
	Func      EndpointFunc[Req, Res]
	Validator Validator
}

func NewEndpoint[Req, Res any](
	endpointFunc EndpointFunc[Req, Res],
) (ep *Endpoint[Req, Res]) {
	ep = &Endpoint[Req, Res]{
		Func: endpointFunc,
	}
	return
}

func NewRawEndpoint(
	endpointFunc HandlerFunc,
) (ep *Endpoint[any, any]) {
	ep = &Endpoint[any, any]{
		Func: func(ctx context.Context, request any) (response any, err error) {
			return nil, endpointFunc(ctx.(Context))
		},
	}
	return
}

func (endpoint Endpoint[Req, Res]) AsHandler() Handler {
	return endpoint
}

func (endpoint Endpoint[Req, Res]) WithValidator(validator Validator) Endpoint[Req, Res] {
	endpoint.Validator = validator
	return endpoint
}

func (endpoint Endpoint[Req, Res]) WithMiddleware(middleware ...Middleware) Handler {
	return ApplyMiddleware(endpoint.AsHandler(), middleware...)
}

// Serve implements transport.Handler
// TODO: benchmark value receiver vs pointer receiver (maybe have request overhead)
func (endpoint Endpoint[Req, Res]) Serve(ctx Context) (err error) {
	decoded := new(Req)

	err = ctx.Codec().Decode(ctx, ctx.Request(), decoded)
	if err != nil {
		return ctx.Codec().EncodeError(ctx, ctx.Response(), err)
	}

	if endpoint.Validator != nil {
		if err = endpoint.Validator.Validate(ctx, decoded); err != nil {
			return ctx.Codec().EncodeError(ctx, ctx.Response(), err)
		}
	}

	if v, ok := asAny(decoded).(PayloadValidator); ok {
		if err = v.Validate(); err != nil {
			return ctx.Codec().EncodeError(ctx, ctx.Response(), err)
		}
	}

	response, err := endpoint.Func(ctx, *decoded)
	if err != nil {
		return ctx.Codec().EncodeError(ctx, ctx.Response(), err)
	}

	return ctx.Codec().Encode(ctx, ctx.Response(), response)
}

func asAny[T any](t *T) any {
	return t
}
