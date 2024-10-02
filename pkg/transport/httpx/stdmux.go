package httpx

import (
	"net/http"
)

type ServeMux interface {
	http.Handler

	// Handle registers a Handler with its enclosed muxer
	// If a handler already exists for pattern it may panic.
	Handle(handler *Handler)
}

var _ ServeMux = (*StdLibMux)(nil)

type StdLibMux struct {
	mux *http.ServeMux
}

func (s StdLibMux) Handle(handler *Handler) {
	s.mux.Handle(handler.Path, handler)
}

func NewStdLibMux() *StdLibMux {
	return &StdLibMux{
		mux: http.NewServeMux(),
	}
}

func (s StdLibMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
