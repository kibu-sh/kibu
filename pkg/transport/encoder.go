package transport

import (
	"context"
	"encoding/json"
)

// Encoder is a function that transforms a concrete response type from an EndpointFunc into a raw byte stream.
// Typically, this is used to serialize data raw byte data like JSON or Protobuf.
// This can also be used to map data back on a transport specific response (i.e. headers, cookies, etc).
type Encoder interface {
	Encode(ctx context.Context, writer ResponseWriter, response any) error
}

// EncoderFunc is a functional implementation to the Encoder interface
type EncoderFunc func(ctx context.Context, writer ResponseWriter, response any) error

// Encode implements Encoder
func (e EncoderFunc) Encode(ctx context.Context, writer ResponseWriter, response any) error {
	return e(ctx, writer, response)
}

// ErrorEncoder encodes an error uses ResponseWriter to send the error to the client
type ErrorEncoder interface {
	Encode(ctx context.Context, writer ResponseWriter, err error) error
}

// ErrorEncoderFunc is a functional implementation to the ErrorEncoder interface
type ErrorEncoderFunc func(ctx context.Context, writer ResponseWriter, err error) error

// Encode implements ErrorEncoder
func (e ErrorEncoderFunc) Encode(ctx context.Context, writer ResponseWriter, err error) error {
	return e(ctx, writer, err)
}

// JSONEncoder encodes any response as JSON and writes it to the ResponseWriter
func JSONEncoder() EncoderFunc {
	return func(ctx context.Context, writer ResponseWriter, response any) error {
		return json.NewEncoder(writer).Encode(response)
	}
}

// JSONErrorEncoder encodes any response as JSON and writes it to the ResponseWriter
func JSONErrorEncoder() ErrorEncoderFunc {
	return func(ctx context.Context, writer ResponseWriter, err error) error {
		return json.NewEncoder(writer).Encode(err)
	}
}
