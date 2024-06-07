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
	Func       EndpointFunc[Req, Res]
	Validator  Validator
	Middleware []Middleware
}

func NewEndpoint[Req, Res any](
	endpointFunc EndpointFunc[Req, Res],
) (ep *Endpoint[Req, Res]) {
	ep = &Endpoint[Req, Res]{
		Func: endpointFunc,
	}
	return
}

func (endpoint Endpoint[Req, Res]) WithValidator(validator Validator) Endpoint[Req, Res] {
	endpoint.Validator = validator
	return endpoint
}

func (endpoint Endpoint[Req, Res]) WithMiddleware(middleware ...Middleware) Endpoint[Req, Res] {
	endpoint.Middleware = middleware
	return endpoint
}

// Serve implements transport.Handler
// TODO: benchmark value receiver vs pointer receiver (maybe have request overhead)
func (endpoint Endpoint[Req, Res]) Serve(tctx Context) (err error) {
	decoded := new(Req)
	codec := tctx.Codec()
	rawReq := tctx.Request()
	rawRes := tctx.Response()
	rawCtx := tctx.Request().Context()

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

	response, err := endpoint.execute(tctx, *decoded)
	if errors.Is(err, ErrResponseIntercepted) {
		err = nil
		return
	}

	if err != nil {
		return codec.EncodeError(rawCtx, rawRes, err)
	}

	return tctx.Codec().Encode(rawCtx, rawRes, response)
}

// execute applies all middleware before execution of the primary endpoint.Func and captures the response
func (endpoint Endpoint[Req, Res]) execute(tctx Context, req Req) (res Res, err error) {
	err = ApplyMiddleware(endpoint.asHandlerWithRespCapture(req, &res), endpoint.Middleware...).Serve(tctx)
	return
}

// asHandlerWithRespCapture converts the endpoint func into a HandlerFunc
// the response pointer is overwritten when the HandlerFunc is executed
func (endpoint Endpoint[Req, Res]) asHandlerWithRespCapture(req Req, res *Res) HandlerFunc {
	return func(tctx Context) (err error) {
		// allows endpoint to access the original transport context with a signature of context.Context
		envelopedTransportCtx := ContextStore.Save(tctx.Request().Context(), tctx)
		*res, err = endpoint.Func(envelopedTransportCtx, req)
		return
	}
}

func asAny[T any](t *T) any {
	return t
}
