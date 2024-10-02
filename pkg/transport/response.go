package transport

import (
	"bytes"
	"io"
	"net/http"
)

// Response is a transport agnostic interface that maps data to a connection
type Response interface {
	io.Writer

	// Headers return a set of key value pairs that represent the headers of the underlying transport destination
	// some transports like Kafka, NATs, and temporal share similar header semantics to HTTP
	Headers() http.Header

	// SetStatusCode sets the status code of the response
	SetStatusCode(int)

	// GetStatusCode returns the status code sent to the client of the response
	GetStatusCode() int

	// BytesWritten returns the number of bytes written to the transport destination
	BytesWritten() int64

	// DelCookie deletes a cookie from the response
	DelCookie(cookie http.Cookie) Response

	// DelCookieByName deletes a cookie from the response by name
	DelCookieByName(name string) Response

	// SetCookie sets a cookie on the response
	SetCookie(cookie http.Cookie) Response

	// Redirect redirects the response to a new url with a given status code
	// specifically useful over HTTP using the Location header
	Redirect(req Request, url string, code int)

	// BodyBuffer returns a buffer that can be used to read the body of the response
	// this should only be used after the response has been written
	// once the response has been written; the buffer will contain the bytes written to the response
	BodyBuffer() *bytes.Buffer

	// Underlying returns a transport-specific response
	// it should return an interface or pointer to the original response (i.e. http.ResponseWriter)
	// this should be used with care as it couples your code to a specific transport
	// this is only provided for break glass scenarios where you need raw access
	Underlying() any
}
