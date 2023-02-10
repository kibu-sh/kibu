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

type errResponse struct {
	Message string `json:"message"`
}

func (e errResponse) ErrResponse() any {
	return e
}

type ErrStatus interface {
	Status() int
}

type ErrResponse interface {
	ErrResponse() any
}

// JSONErrorEncoder encodes any response as JSON and writes it to the Response
func JSONErrorEncoder() transport.ErrorEncoderFunc {
	return func(ctx context.Context, writer transport.Response, err error) error {
		// TODO: inherit status code from err if type is *ApiError
		var res ErrResponse = &errResponse{
			Message: err.Error(),
		}

		status := http.StatusInternalServerError
		if err, ok := err.(ErrStatus); ok {
			status = err.Status()
		}

		// prefer custom error response
		if errRes, ok := err.(ErrResponse); ok {
			res = errRes
		}

		writer.SetStatusCode(status)
		writer.Headers().Set("Content-Type", "application/json")
		return json.NewEncoder(writer).Encode(res)
	}
}
