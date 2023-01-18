package transport

import (
	"io"
	"net/http"
)

// Response is a transport agnostic interface that maps data to a connection
type Response interface {
	io.Writer

	Headers() http.Header

	// Underlying returns a transport specific response
	// it should return an interface or pointer to the original response (i.e. http.ResponseWriter)
	// this should be used with care as it couples your code to a specific transport
	// this is only provided for break glass scenarios where you need raw access
	Underlying() any
}
