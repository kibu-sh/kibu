package transport

import (
	"context"
)

// Endpoint is any function that can be modeled as an RPC call.
// These should remain transport agnostic and are used to implement business logic.
type Endpoint[Req, Res any] func(ctx context.Context, req Req) (res Res, err error)

// Middleware are function that decorate endpoints with additional functionality.
// Examples of middleware would be logging, tracing, authorization, etc.
type Middleware[Req, Res any] func(next Endpoint[Req, Res]) Endpoint[Req, Res]

// Decoder is a function that decodes a transport specific request into a concrete type accepted by an Endpoint.
// Use these functions to bind protobuf, json, xml, query params, headers, and cookies to your Endpoint.
type Decoder[T any] func(ctx context.Context, input *T) error

// DecoderChain is a function that decorates a Decoder.
// Use this to create a decoding pipeline.
type DecoderChain[T any] func(next Decoder[T]) Decoder[T]

// Encoder is a function that transforms a concrete response type from an Endpoint into a raw byte stream.
// Typically, this is used to serialize data raw byte data like JSON or Protobuf.
// This can also be used to map data back on a transport specific response (i.e. headers, cookies, etc).
type Encoder[T any] func(ctx context.Context, output *T) error

// Controller is a transport agnostic interface that can be used to implement a server.
type Controller interface {
	Serve(ctx context.Context) error
}
