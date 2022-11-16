package transport

import (
	"context"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"
	"mime"
	"net/http"
)

// Decoder is a function that decodes a transport specific request into a concrete type accepted by an EndpointFunc.
// Use these functions to bind protobuf, json, xml, query params, headers, and cookies to your EndpointFunc.
type Decoder interface {
	Decode(ctx context.Context, request Request, target any) (err error)
}

// DecoderFunc is a functional implementation to the Decoder interface
type DecoderFunc func(ctx context.Context, request Request, target any) (err error)

// DecoderMiddlewareFunc chains Decoders together for multiple decoding passes
// This is useful when parts of your request data are found in different places
// For example, you may want to decode a request body into a struct and then decode a query param into the same struct
// Additionally, you can check for expected content types
type DecoderMiddlewareFunc func(next Decoder) Decoder

// Decode implements Decoder
func (d DecoderFunc) Decode(ctx context.Context, request Request, target any) (err error) {
	return d(ctx, request, target)
}

func HyperMediaDecoder() DecoderFunc {
	return func(ctx context.Context, request Request, target any) (err error) {
		r, ok := request.Underlying().(*http.Request)
		if !ok {
			return
		}

		contentType, _, err := request.ParseMediaType()
		if errors.Is(err, mime.ErrInvalidMediaParameter) {
			return err
		}

		return binding.Default(request.Method(), contentType).Bind(r, target)
	}
}

type ValueSet map[string][]string

func ValueDecoder(tag string, get func(request Request) ValueSet) DecoderFunc {
	return func(ctx context.Context, request Request, target any) (err error) {
		if params := get(request); params != nil {
			err = binding.MapFormWithTag(target, params, tag)
		}
		return
	}
}

func PathParamsDecoder() DecoderFunc {
	return ValueDecoder("path", func(request Request) ValueSet {
		return ValueSet(request.PathParams())
	})
}

func QueryParamsDecoder() DecoderFunc {
	return ValueDecoder("query", func(request Request) ValueSet {
		return ValueSet(request.QueryParams())
	})
}

func HeaderDecoder() DecoderFunc {
	return ValueDecoder("header", func(request Request) ValueSet {
		return ValueSet(request.Headers())
	})
}

func CookieDecoder() DecoderFunc {
	return ValueDecoder("cookie", func(request Request) (result ValueSet) {
		result = make(ValueSet)
		for _, cookie := range request.Cookies() {
			result[cookie.Name] = append(result[cookie.Name], cookie.Value)
		}
		return
	})
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

func DefaultDecoderChain() DecoderFunc {
	return NewDecoderChain(
		PathParamsDecoder(),
		QueryParamsDecoder(),
		HeaderDecoder(),
		CookieDecoder(),
		HyperMediaDecoder(),
	)
}
