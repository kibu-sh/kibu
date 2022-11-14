package httpx

import (
	"context"
	"encoding/json"
	"github.com/discernhq/devx/pkg/transport"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"
	"mime"
	"net/http"
	"net/url"
)

var uriParamsContextKey struct{}

// PathParamsFromContext TODO: write test
func PathParamsFromContext(ctx context.Context) url.Values {
	if v, ok := ctx.Value(uriParamsContextKey).(url.Values); ok {
		return v
	}
	return nil
}

// ContextWithPathParams TODO: implement this as middleware for Gin or what ever Router
func ContextWithPathParams(ctx context.Context, values url.Values) context.Context {
	return context.WithValue(ctx, uriParamsContextKey, values)
}

type DecoderFunc[T any] func(r *http.Request) transport.Decoder[T]
type EncoderFunc[T any] func(w http.ResponseWriter) transport.Encoder[T]

// Decode TODO: break this into smaller decoders
func Decode[T any](r *http.Request) transport.Decoder[T] {
	return func(ctx context.Context, input *T) (err error) {
		if params := PathParamsFromContext(ctx); params != nil {
			if err = binding.MapFormWithTag(input, params, "path"); err != nil {
				return
			}
		}

		m, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if errors.Is(err, mime.ErrInvalidMediaParameter) {
			return err
		}

		if err = binding.Default(r.Method, m).Bind(r, input); err != nil {
			return
		}

		if err = binding.MapFormWithTag(input, r.URL.Query(), "query"); err != nil {
			return
		}

		if err = binding.Header.Bind(r, input); err != nil {
			return
		}

		if err = binding.Cookie.Bind(r, input); err != nil {
			return
		}

		return nil
	}
}

func Encode[T any](w http.ResponseWriter) transport.Encoder[T] {
	return func(ctx context.Context, output *T) (err error) {
		return json.NewEncoder(w).Encode(output)
	}
}
