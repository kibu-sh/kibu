package transport

import (
	"context"
)

// Decoder is a function that decodes a transport specific request into a concrete type accepted by an EndpointFunc.
// Use these functions to bind protobuf, json, xml, query params, headers, and cookies to your EndpointFunc.
type Decoder interface {
	Decode(ctx context.Context, request Request, target any) (err error)
}

// DecoderFunc is a functional implementation to the Decoder interface
type DecoderFunc func(ctx context.Context, request Request, target any) (err error)

// Decode implements Decoder
func (d DecoderFunc) Decode(ctx context.Context, request Request, target any) (err error) {
	return d(ctx, request, target)
}

func NewDecoderChain(decoders ...Decoder) DecoderFunc {
	return func(ctx context.Context, request Request, target any) (err error) {
		for _, decoder := range decoders {
			if err = decoder.Decode(ctx, request, target); err != nil {
				return
			}
		}
		return
	}
}
