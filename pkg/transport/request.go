package transport

import (
	"io"
	"net/http"
	"net/url"
)

// Request is a transport agnostic interface that maps data from a connection
type Request interface {
	URL() *url.URL
	Path() string
	Method() string
	PathParams() url.Values
	QueryParams() url.Values
	Headers() http.Header
	Cookies() []*http.Cookie

	// Body exposes io.ReadCloser from the Underlying request
	// We recommend using http.MaxBytesReader to limit the size of the body
	// Alternatively you can use io.LimitReader to limit the size of the body
	// Consider using a middleware function to limit the maximum size of the body
	Body() io.ReadCloser

	// ParseMediaType should forward the return value of mime.ParseMediaType
	ParseMediaType() (mediatype string, params map[string]string, err error)

	// Underlying returns a transport specific request
	// it should return a pointer to the original request (i.e. *http.Request)
	// this should be used with care as it couples your code to a specific transport
	// this is only provided for break glass scenarios where you need raw access
	Underlying() any
}
