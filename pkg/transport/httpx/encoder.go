package httpx

import (
	"context"
	"encoding/json"
	"github.com/discernhq/devx/pkg/transport"
	"net/http"
)

// JSONEncoder encodes any response as JSON and writes it to the ResponseWriter
func JSONEncoder() transport.EncoderFunc {
	return func(ctx context.Context, writer transport.Response, response any) error {
		writer.Headers().Set("Content-Type", "application/json")
		return json.NewEncoder(writer).Encode(response)
	}
}

type DefaultJSONError struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (d DefaultJSONError) GetStatusCode() int {
	return d.Status
}

func (d DefaultJSONError) PrepareResponse() any {
	return d
}

// JSONErrorEncoder encodes any response as JSON and writes it to the ResponseWriter
func JSONErrorEncoder() transport.ErrorEncoderFunc {
	return func(ctx context.Context, writer transport.Response, err error) error {
		res, ok := err.(transport.ErrorResponse)
		if !ok {
			res = DefaultJSONError{
				Status:  http.StatusInternalServerError,
				Message: err.Error(),
			}
		}

		writer.SetStatusCode(res.GetStatusCode())
		writer.Headers().Set("Content-Type", "application/json")
		return json.NewEncoder(writer).Encode(res.PrepareResponse())
	}
}
