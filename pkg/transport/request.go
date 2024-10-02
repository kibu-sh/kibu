package transport

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
)

// Request is a transport agnostic interface that maps data from a connection
type Request interface {
	URL() *url.URL
	Path() string
	Method() string

	Version() string
	PathParams() url.Values
	QueryParams() url.Values
	Headers() http.Header
	Context() context.Context
	Cookies() []*http.Cookie
	Cookie(name string) (cookie *http.Cookie, err error)
	WithContext(ctx context.Context) Request

	// Body exposes io.ReadCloser from the Underlying request
	// We recommend using http.MaxBytesReader to limit the size of the body
	// Alternatively you can use io.LimitReader to limit the size of the body
	// Consider using a middleware function to limit the maximum size of the body
	Body() io.ReadCloser

	// BodyBuffer returns a buffer that can be used to read the body of the request
	// This should only be used once the body has been read
	// This contains a copy of all the bytes read from the original body
	BodyBuffer() *bytes.Buffer

	// ParseMediaType should forward the return value of mime.ParseMediaType
	ParseMediaType() (mediatype string, params map[string]string, err error)

	// Underlying returns a transport specific request
	// it should return a pointer to the original request (i.e. *http.Request)
	// this should be used with care as it couples your code to a specific transport
	// this is only provided for break glass scenarios where you need raw access
	Underlying() any
}
