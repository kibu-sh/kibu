package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var (
	ErrStatusCheckFailed  = errors.New("unable to send http request")
	ErrInvalidHTTPRequest = errors.New("invalid http request")
)

type (
	StatusCheckFunc   func(status string, code int) error
	JSONRequestParams struct {
		Method      string
		Client      *http.Client
		Url         *url.URL
		Body        any
		Headers     http.Header
		StatusCheck StatusCheckFunc
	}

	ErrorResponse struct {
		URL        *url.URL      `json:"url"`
		Method     string        `json:"method"`
		StatusCode int           `json:"status_code"`
		Body       *bytes.Buffer `json:"body"`
		Header     http.Header   `json:"header"`
		Err        error         `json:"-"`
	}
)

func (r *ErrorResponse) Wrap(err error) error {
	return &ErrorResponse{
		URL:        r.URL,
		StatusCode: r.StatusCode,
		Body:       r.Body,
		Header:     r.Header,
		Err:        errors.Join(r, err),
	}
}

func (r *ErrorResponse) Unwrap() error {
	return r.Err
}

var errResponseTemplate = `
http request failed: 
method: %s
url: %s
status: %d
error: %v
body: %s
header: %s
`

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf(
		errResponseTemplate,
		r.Method,
		r.URL,
		r.StatusCode,
		r.Err,
		r.Body,
		headerAsString(r.Header),
	)
}

func headerAsString(header http.Header) string {
	var buf bytes.Buffer
	buf.WriteString("\n")
	for k, v := range header {
		buf.WriteString(fmt.Sprintf("\t%s: %s\n", k, v))
	}
	return buf.String()
}

// NewErrorResponse attempts to read the response body after a request
// This is to improve visibility into HTTP request failures
func NewErrorResponse(res *http.Response) *ErrorResponse {
	var err error
	buf := new(bytes.Buffer)
	if res.Body != http.NoBody {
		_, err = io.Copy(buf, res.Body)
	}
	return &ErrorResponse{
		URL:        res.Request.URL,
		Method:     res.Request.Method,
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       buf,
		Err:        err,
	}
}

// JSONRequest function is a generic HTTP client function that can send HTTP requests and decode responses into any Go struct or value.
// The function takes a context, HTTP method, URL, request body, and a map of headers as inputs.
// It uses the default HTTP client provided by Go to make the HTTP request, with the request body and headers added to the request
func JSONRequest[T any](ctx context.Context, params JSONRequestParams) (result T, err error) {
	client := useDefaultClientAsFallback(params)
	statusCheck := useDefaultStatusCheckAsFallback(params)
	payload, err := encodeOptionalPayloadAsJSON(params.Body)
	if err != nil {
		return
	}

	req, err := newRequestWithContextAndHeaders(
		ctx,
		params.Method,
		params.Url,
		payload,
		params.Headers,
	)
	if err != nil {
		return
	}

	res, err := client.Do(req)
	defer func() {
		if err != nil {
			err = NewErrorResponse(res).Wrap(err)
		}
		_ = res.Body.Close()
	}()

	if err != nil {
		return
	}

	if err = statusCheck(res.Status, res.StatusCode); err != nil {
		return
	}

	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return
	}
	return
}

func encodeOptionalPayloadAsJSON(body any) (io.Reader, error) {
	if body == nil {
		return http.NoBody, nil
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return nil, err
	}
	return buf, nil
}

func newRequestWithContextAndHeaders(
	ctx context.Context,
	method string,
	url *url.URL,
	body io.Reader,
	headers http.Header,
) (req *http.Request, err error) {
	req, err = http.NewRequestWithContext(ctx, method, url.String(), body)
	if err != nil {
		err = errors.Join(ErrInvalidHTTPRequest, err)
		return
	}
	req.Header = headers
	return
}

func useDefaultStatusCheckAsFallback(params JSONRequestParams) StatusCheckFunc {
	if params.StatusCheck == nil {
		return DefaultStatusCheckFunc
	}
	return params.StatusCheck
}

func useDefaultClientAsFallback(params JSONRequestParams) *http.Client {
	if params.Client == nil {
		return http.DefaultClient
	}
	return params.Client
}

func DefaultStatusCheckFunc(status string, code int) (err error) {
	if code != http.StatusOK {
		err = errors.Join(
			ErrStatusCheckFailed,
			fmt.Errorf("status:%s cod: %d", status, code),
		)
	}
	return
}
