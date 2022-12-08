package httpx

import (
	"context"
	"github.com/samber/lo"
	"net"
	"net/http"
	"time"
)

type Server struct {
	*http.Server
	Listener        net.Listener
	shutdownTimeout *time.Duration
	shutdownChan    chan error
}

func (s *Server) Start(ctx context.Context) error {
	startChan := make(chan error)

	go func() {
		close(startChan)
		_ = s.Serve(s.Listener)
	}()

	go func() {
		<-ctx.Done()
		sCtx, cancel := context.WithTimeout(context.Background(), *s.shutdownTimeout)
		defer cancel()
		defer close(s.shutdownChan)
		s.shutdownChan <- s.Shutdown(sCtx)
	}()

	return <-startChan
}

func (s *Server) Stop(ctx context.Context) error {
	return s.Shutdown(ctx)
}

func (s *Server) Wait() error {
	// TODO: think about our timeouts
	return <-s.shutdownChan
}

type NewServerParams struct {
	Addr     ListenAddr
	Mux      ServeMux
	Handlers []*Handler
}

type ListenAddr string

func NewServer(params *NewServerParams) (*Server, error) {
	listener, err := net.Listen("tcp", string(params.Addr))
	if err != nil {
		return nil, err
	}

	// TODO: register global middleware
	// register handlers with mux router
	for _, handler := range params.Handlers {
		params.Mux.Handle(handler)
	}

	return &Server{
		Server: &http.Server{
			Addr:    string(params.Addr),
			Handler: params.Mux,
		},
		Listener: listener,

		// buffered to avoid blocking on shutdown
		shutdownChan:    make(chan error, 1),
		shutdownTimeout: lo.ToPtr(time.Second * 30),
	}, nil
}
