package transport

import (
	"context"
	"github.com/pkg/errors"
)

// ErrResponseIntercepted should be returned by any endpoint that wants to overload the default response behavior
var ErrResponseIntercepted = errors.New("raw response")

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
		Func: newRawEndpointFunc(endpointFunc),
	}
	return
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
	codec := ctx.Codec()
	rawReq := ctx.Request()
	rawRes := ctx.Response()
	rawCtx := ctx.Request().Context()
	// allows upstream middleware to access the transport context even when it's masquerading as a context.Context
	envelopedTransportCtx := ContextStore.Save(rawCtx, ctx)

	if err = codec.Decode(rawCtx, rawReq, decoded); err != nil {
		return codec.EncodeError(rawCtx, rawRes, err)
	}

	if endpoint.Validator != nil {
		if err = endpoint.Validator.Validate(rawCtx, decoded); err != nil {
			return codec.EncodeError(rawCtx, rawRes, err)
		}
	}

	if v, ok := asAny(decoded).(PayloadValidator); ok {
		if err = v.Validate(); err != nil {
			return codec.EncodeError(rawCtx, rawRes, err)
		}
	}

	response, err := endpoint.Func(envelopedTransportCtx, *decoded)
	if errors.Is(err, ErrResponseIntercepted) {
		err = nil
		return
	}

	if err != nil {
		return codec.EncodeError(rawCtx, rawRes, err)
	}

	return ctx.Codec().Encode(rawCtx, rawRes, response)
}

func asAny[T any](t *T) any {
	return t
}
