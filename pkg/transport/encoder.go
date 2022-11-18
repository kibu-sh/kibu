package transport

import (
	"context"
)

// TODO: should encoder use transport.Context?

// Encoder is a function that transforms a concrete response type from an EndpointFunc into a raw byte stream.
// Typically, this is used to serialize data raw byte data like JSON or Protobuf.
// This can also be used to map data back on a transport specific response (i.e. headers, cookies, etc).
type Encoder interface {
	Encode(ctx context.Context, writer Response, response any) error
}

// EncoderFunc is a functional implementation to the Encoder interface
type EncoderFunc func(ctx context.Context, writer Response, response any) error

// Encode implements Encoder
func (e EncoderFunc) Encode(ctx context.Context, writer Response, response any) error {
	return e(ctx, writer, response)
}

// ErrorEncoder encodes an error uses Response to send the error to the client
type ErrorEncoder interface {
	EncodeError(ctx context.Context, writer Response, err error) error
}

// ErrorEncoderFunc is a functional implementation to the ErrorEncoder interface
type ErrorEncoderFunc func(ctx context.Context, writer Response, err error) error

// EncodeError implements ErrorEncoder
func (e ErrorEncoderFunc) EncodeError(ctx context.Context, writer Response, err error) error {
	return e(ctx, writer, err)
}
