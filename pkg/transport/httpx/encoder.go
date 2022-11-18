package httpx

import (
	"context"
	"encoding/json"
	"github.com/discernhq/devx/pkg/transport"
)

// JSONEncoder encodes any response as JSON and writes it to the Response
func JSONEncoder() transport.EncoderFunc {
	return func(ctx context.Context, writer transport.Response, response any) error {
		return json.NewEncoder(writer).Encode(response)
	}
}

// JSONErrorEncoder encodes any response as JSON and writes it to the Response
func JSONErrorEncoder() transport.ErrorEncoderFunc {
	return func(ctx context.Context, writer transport.Response, err error) error {
		return json.NewEncoder(writer).Encode(err)
	}
}
