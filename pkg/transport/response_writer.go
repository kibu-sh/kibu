package transport

import "io"

// ResponseWriter is a transport agnostic interface that maps data to a connection
type ResponseWriter interface {
	io.Writer
	// Underlying returns a transport specific response
	// it should return an interface or pointer to the original response (i.e. http.ResponseWriter)
	Underlying() any
}
