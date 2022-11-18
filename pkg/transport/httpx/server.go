package httpx

import (
	"context"
	"github.com/discernhq/devx/pkg/transport"
	"net"
	"net/http"
)

type Server struct {
	*http.Server
	l net.Listener
}

func (s *Server) Start(ctx context.Context) error {
	return s.Serve(s.l)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.Shutdown(ctx)
}

func OnStart(ctx context.Context, s transport.Server) error {
	// TODO: this might not be super great
	// lets think about what happens when this errors
	go s.Start(ctx)
	return nil
}

func OnStop(ctx context.Context, s transport.Server) error {
	return s.Stop(ctx)
}

type NewServerParams struct {
	Addr string
	Mux  ServeMux
}

func NewServer(params *NewServerParams) (transport.Server, error) {
	s := &http.Server{
		Addr:    params.Addr,
		Handler: params.Mux,
	}

	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return nil, err
	}

	return &Server{
		s, l,
	}, nil
}
