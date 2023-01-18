package httpx

import (
	"context"
	"encoding/json"
	"github.com/discernhq/devx/pkg/transport"
	"net/http"
)

// JSONEncoder encodes any response as JSON and writes it to the Response
func JSONEncoder() transport.EncoderFunc {
	return func(ctx context.Context, writer transport.Response, response any) error {
		writer.Headers().Set("Content-Type", "application/json")
		return json.NewEncoder(writer).Encode(response)
	}
}

// TODO: define concrete API error types

type ErrorResponse struct {
	Message string `json:"message"`
}

// JSONErrorEncoder encodes any response as JSON and writes it to the Response
func JSONErrorEncoder() transport.ErrorEncoderFunc {
	return func(ctx context.Context, writer transport.Response, err error) error {
		// TODO: inherit status code from err if type is *ApiError
		writer.SetStatusCode(http.StatusInternalServerError)
		writer.Headers().Set("Content-Type", "application/json")
		return json.NewEncoder(writer).Encode(&ErrorResponse{
			Message: err.Error(),
		})
	}
}
