package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func (d DefaultJSONError) Error() string {
	return fmt.Sprintf("%d: %s", d.Status, d.Message)
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
		var res transport.ErrorResponse = DefaultJSONError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
		_ = errors.As(err, &res)

		writer.SetStatusCode(res.GetStatusCode())
		writer.Headers().Set("Content-Type", "application/json")
		return json.NewEncoder(writer).Encode(res.PrepareResponse())
	}
}
