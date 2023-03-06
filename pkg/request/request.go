package request

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

// ErrStatusCheckFailed is returned when the status check function returns an error.
// This enables the user to determine if the error is a status check failure.
var ErrStatusCheckFailed = errors.New("status check failed")

// Client is a wrapper around http.Client that provides a fluent interface for making requests.
type Client struct {
	c               *http.Client
	baseURL         *url.URL
	statusCheckFunc StatusCheckFunc
	defaultHeader   http.Header
	method          string
	body            io.Reader
	deferredBody    func() (io.Reader, error)
}

// WithStatusCheckFunc returns a new instance of Client by replacing its status check function.
func (c Client) WithStatusCheckFunc(f StatusCheckFunc) *Client {
	c.statusCheckFunc = f
	return &c
}

// WithBaseURL returns a new instance of Client by replacing its base URL.
func (c Client) WithBaseURL(url *url.URL) *Client {
	c.baseURL = url
	return &c
}

// WithJoinedURLPath returns a new instance of Client with an updated base URL.
// The supplied path is joined to the base URL.
// A baseURL of "http://test.com" using WithJoinedURLPath("/foo") will produce a new URL of "http://test.com/foo"
func (c Client) WithJoinedURLPath(path string) *Client {
	c.baseURL = c.baseURL.JoinPath(path)
	return &c
}

// WithClient returns a new instance of Client by replacing its http.Client.
func (c Client) WithClient(client *http.Client) *Client {
	c.c = client
	return &c
}

// WithRoundTripper returns a new instance of Client with an updated http.RoundTripper.
// This is useful for intercepting requests and responses.
func (c Client) WithRoundTripper(rt http.RoundTripper) *Client {
	c.c.Transport = rt
	return &c
}

// WithDefaultHeader returns a new instance of Client by replacing its default header
func (c Client) WithDefaultHeader(header http.Header) *Client {
	c.defaultHeader = header
	return &c
}

// WithHeader returns a new instance of Client by cloning its default header and appending the key/value pair.
// If an existing key is found, the value is appended to the existing value.
func (c Client) WithHeader(key, value string) *Client {
	c.defaultHeader = c.defaultHeader.Clone()
	c.defaultHeader.Add(key, value)
	return &c
}

// WithMethod returns a new instance of Client by replacing its HTTP method.
func (c Client) WithMethod(method string) *Client {
	c.method = method
	return &c
}

// WithGet returns a new instance of Client with an HTTP method of GET.
func (c Client) WithGet() *Client {
	return c.WithMethod(http.MethodGet)
}

// WithPost returns a new instance of Client with an HTTP method of POST.
func (c Client) WithPost() *Client {
	return c.WithMethod(http.MethodPost)
}

// WithPut returns a new instance of Client with an HTTP method of PUT.
func (c Client) WithPut() *Client {
	return c.WithMethod(http.MethodPut)
}

// WithPatch returns a new instance of Client with an HTTP method of PATCH.
func (c Client) WithPatch() *Client {
	return c.WithMethod(http.MethodPatch)
}

// WithDelete returns a new instance of Client with an HTTP method of DELETE.
func (c Client) WithDelete() *Client {
	return c.WithMethod(http.MethodDelete)
}

// WithHead returns a new instance of Client with an HTTP method of HEAD.
func (c Client) WithHead() *Client {
	return c.WithMethod(http.MethodHead)
}

// WithOptions returns a new instance of Client with an HTTP method of OPTIONS.
func (c Client) WithOptions() *Client {
	return c.WithMethod(http.MethodOptions)
}

// WithTrace returns a new instance of Client with an HTTP method of TRACE.
func (c Client) WithTrace() *Client {
	return c.WithMethod(http.MethodTrace)
}

// WithConnect returns a new instance of Client with an HTTP method of CONNECT.
func (c Client) WithConnect() *Client {
	return c.WithMethod(http.MethodConnect)
}

// WithBody returns a new instance of Client by replacing its body.
func (c Client) WithBody(body io.Reader) *Client {
	c.body = body
	return &c
}

// WithExpectStatus returns a new instance of Client by replacing its status check function.
// The supplied code must exactly match the response status code.
func (c Client) WithExpectStatus(code int) *Client {
	return c.WithStatusCheckFunc(NewExactStatusCheckFunc(code))
}

// WithExpectStatusInRange returns a new instance of Client by replacing its status check function.
// The supplied code must be within the specified range matching the response status code.
func (c Client) WithExpectStatusInRange(min, max int) *Client {
	return c.WithStatusCheckFunc(NewStatusRangeCheckFunc(min, max))
}

// WithDeferredBody returns a new instance of Client by replacing its deferrable body.
// The supplied function isn't executed until the request is made.
// This is preferred to the default `
func (c Client) WithDeferredBody(f func() (io.Reader, error)) *Client {
	c.deferredBody = f
	return &c
}

// WithJSONBody returns a new instance of Client by replacing its deferred body.
// The supplied body is encoded as JSON before the request is sent.
func (c Client) WithJSONBody(body any) *Client {
	return c.
		WithHeader("Content-Type", "application/json").
		WithDeferredBody(func() (io.Reader, error) {
			buf := new(bytes.Buffer)
			err := json.NewEncoder(buf).Encode(body)
			return buf, err
		})
}

// Do execute the request, checks the status code, and returns the response.
func (c Client) Do(ctx context.Context) (res *http.Response, err error) {
	body := c.body
	if c.deferredBody != nil {
		if body, err = c.deferredBody(); err != nil {
			return
		}
	}

	req, err := http.NewRequestWithContext(
		ctx,
		c.method,
		c.baseURL.String(),
		body,
	)
	if err != nil {
		return
	}

	req.Header = c.defaultHeader

	res, err = c.c.Do(req)
	if err != nil {
		return
	}

	if c.statusCheckFunc != nil {
		if err = c.statusCheckFunc(res.Status, res.StatusCode); err != nil {
			return
		}
	}

	return
}

// DoAsJSON executes the request, checks the status code, and decodes the response body as JSON.
func (c Client) DoAsJSON(ctx context.Context, result any) (err error) {
	res, err := c.Do(ctx)
	if err != nil {
		return
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if err = json.NewDecoder(res.Body).Decode(result); err != nil {
		return
	}
	return
}

// NewClient returns a new instance of Client with default options.
func NewClient(baseURL *url.URL) *Client {
	return (&Client{}).
		WithBaseURL(baseURL).
		WithBody(http.NoBody).
		WithClient(http.DefaultClient).
		WithDefaultHeader(http.Header{}).
		WithStatusCheckFunc(NewOkayRangeCheckFunc())
}

// ParseURL parses the supplied string as a URL and returns a new instance of Client with default options.
func ParseURL(s string) (*Client, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	return NewClient(u), nil
}
