package transport

import (
	"context"
	"github.com/pkg/errors"
)

// Controller processes transport.Request and binds it to an Endpoint
type Controller struct {
	Endpoint    Endpoint
	Decoder     Decoder
	Encoder     Encoder
	Validator   Validator
	ErrEncoder  ErrorEncoder
	initRequest func() any
}

type ControllerOption func(ctrl *Controller)

func NewController(
	endpoint Endpoint,
	opts ...ControllerOption,
) (ctrl *Controller) {
	ctrl = &Controller{
		Endpoint:    endpoint,
		Encoder:     JSONEncoder(),
		Decoder:     DefaultDecoderChain(),
		ErrEncoder:  JSONErrorEncoder(),
		initRequest: endpoint.InitRequest,
	}

	for _, opt := range opts {
		opt(ctrl)
	}

	return
}

// WithDecoders sets the Decoder for the Controller
// Overrides DefaultDecoderChain
func WithDecoders(decoders ...Decoder) ControllerOption {
	return func(mux *Controller) {
		mux.Decoder = NewDecoderChain(decoders...)
	}
}

func WithValidator(validator Validator) ControllerOption {
	return func(mux *Controller) {
		mux.Validator = validator
	}
}

// WithMiddleware decorates a Controller.Endpoint
// These will be called in order before Controller.Endpoint
func WithMiddleware(decorators ...EndpointMiddlewareFunc) ControllerOption {
	return func(mux *Controller) {
		if mux.Endpoint == nil {
			panic(errors.New("cannot wrap nil endpoint"))
		}

		for _, decorator := range decorators {
			mux.Endpoint = decorator(mux.Endpoint)
		}
	}
}

// Serve implements transport.Handler
func (s Controller) Serve(ctx context.Context, request Request, writer ResponseWriter) (err error) {
	decoded := s.initRequest()

	err = s.Decoder.Decode(ctx, request, decoded)
	if err != nil {
		return s.ErrEncoder.Encode(ctx, writer, err)
	}

	if s.Validator != nil {
		if err = s.Validator.Validate(ctx, decoded); err != nil {
			return s.ErrEncoder.Encode(ctx, writer, err)
		}
	}

	if v, ok := decoded.(PayloadValidator); ok {
		if err = v.Validate(); err != nil {
			return s.ErrEncoder.Encode(ctx, writer, err)
		}
	}

	response, err := s.Endpoint.Serve(ctx, decoded)
	if err != nil {
		return s.ErrEncoder.Encode(ctx, writer, err)
	}

	return s.Encoder.Encode(ctx, writer, response)
}

// // initializeRequestWithReflection caches the type of the Endpoint's request
// // then uses reflection to initialize a new instance of that type during request processing
// func initializeRequestWithReflection(endpoint Endpoint) func() any {
// 	t := reflect.TypeOf(endpoint.Serve).In(1)
// 	return func() any {
// 		return reflect.New(t).Interface()
// 	}
// }

// // WithTypedRequest overrides the default reflection based request initializer
// // This is useful in performance critical applications
// // Take care as this method isn't actually aware of the endpoints type
// func WithTypedRequest[T any]() ControllerOption {
// 	return func(mux *Controller) error {
// 		mux.initRequest = func() any {
// 			return new(T)
// 		}
// 		return nil
// 	}
// }
