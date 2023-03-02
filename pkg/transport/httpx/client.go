package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
)

// ClientDo  function is a generic HTTP client function that can send HTTP requests and decode responses into any Go struct or value.
// The function takes a context, HTTP method, URL, request body, and a map of headers as inputs.
// It uses the default HTTP client provided by Go to make the HTTP request, with the request body and headers added to the request
var ErrStatusCheckFailed = errors.New("unable to send http request")

type JSONRequestParams struct {
	Method      string
	Client      *http.Client
	Url         *url.URL
	Body        any
	Headers     http.Header
	StatusCheck func(statusCode int, status string) (err error)
}

func JSONRequest[T any](ctx context.Context, params JSONRequestParams) (result T, err error) {

	var req *http.Request
	if params.Client == nil {
		params.Client = http.DefaultClient
	}

	// encode request body
	var payload io.Reader
	if params.Body != nil {
		payload = new(bytes.Buffer)
		if err := json.NewEncoder(payload.(io.Writer)).Encode(params.Body); err != nil {
			return result, err
		}
	}

	// create request
	req, err = http.NewRequestWithContext(ctx, params.Method, params.Url.String(), payload)
	if err != nil {
		err = errors.Wrap(err, "unable to make request")
		return
	}

	// add headers
	req.Header = params.Headers

	res, err := params.Client.Do(req)
	defer func() {
		_ = res.Body.Close()
	}()
	if err != nil {
		return
	}
	if params.StatusCheck == nil {
		params.StatusCheck = DefaultStatusCheckFunc
	}
	err = params.StatusCheck(res.StatusCode, res.Status)
	if err != nil {
		return
	}

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return
	}
	return
}

func DefaultStatusCheckFunc(statusCode int, status string) (err error) {
	if statusCode != http.StatusOK {
		msg := fmt.Sprintf("failed with status:%s:code(%d)", status, statusCode)
		err = errors.Wrap(ErrStatusCheckFailed, msg)
	}
	return
}
