package transport

import (
	"io"
	"net/http"
	"net/url"
)

// Request is a transport agnostic interface that maps data from a connection
type Request interface {
	// Body exposes io.ReadCloser from the Underlying request
	// We recommend using http.MaxBytesReader to limit the size of the body
	// Alternatively you can use io.LimitReader to limit the size of the body
	// Consider using a middleware function to limit the maximum size of the body
	Body() io.ReadCloser

	// ParseMediaType should forward the return value of mime.ParseMediaType
	ParseMediaType() (mediatype string, params map[string]string, err error)

	// Underlying returns a transport specific request
	// it should return a pointer to the original request (i.e. *http.Request)
	Underlying() any
	Method() string
	PathParams() url.Values
	QueryParams() url.Values
	Headers() http.Header
	Cookies() []*http.Cookie
}
