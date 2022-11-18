package httpx

import (
	"context"
	"github.com/discernhq/devx/pkg/transport"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"
	"mime"
	"net/http"
)

type ValueSet map[string][]string

func ValueDecoder(tag string, get func(request transport.Request) ValueSet) transport.DecoderFunc {
	return func(ctx context.Context, request transport.Request, target any) (err error) {
		if params := get(request); params != nil {
			err = binding.MapFormWithTag(target, params, tag)
		}
		return
	}
}

func HyperMediaDecoder() transport.DecoderFunc {
	return func(ctx context.Context, request transport.Request, target any) (err error) {
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

func PathParamsDecoder() transport.DecoderFunc {
	return ValueDecoder("path", func(request transport.Request) ValueSet {
		return ValueSet(request.PathParams())
	})
}

func QueryParamsDecoder() transport.DecoderFunc {
	return ValueDecoder("query", func(request transport.Request) ValueSet {
		return ValueSet(request.QueryParams())
	})
}

func HeaderDecoder() transport.DecoderFunc {
	return ValueDecoder("header", func(request transport.Request) ValueSet {
		return ValueSet(request.Headers())
	})
}

func CookieDecoder() transport.DecoderFunc {
	return ValueDecoder("cookie", func(request transport.Request) (result ValueSet) {
		result = make(ValueSet)
		for _, cookie := range request.Cookies() {
			result[cookie.Name] = append(result[cookie.Name], cookie.Value)
		}
		return
	})
}

func DefaultDecoderChain() transport.Decoder {
	return transport.NewDecoderChain(
		PathParamsDecoder(),
		QueryParamsDecoder(),
		HeaderDecoder(),
		CookieDecoder(),
		HyperMediaDecoder(),
	)
}
