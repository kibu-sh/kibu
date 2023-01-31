package httpx

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"net"
	"net/http"
	"time"
)

type Server struct {
	*http.Server
	Listeners       []net.Listener
	shutdownTimeout *time.Duration
	shutdownChan    chan error
}

func (s *Server) Start(ctx context.Context) error {

	for _, l := range s.Listeners {
		startChan := make(chan error)

		go func(l net.Listener) {
			close(startChan)
			fmt.Println("starting server on", l.Addr().String())
			_ = s.Serve(l)
		}(l)

		go func() {
			<-ctx.Done()
			sCtx, cancel := context.WithTimeout(context.Background(), *s.shutdownTimeout)
			defer cancel()
			defer close(s.shutdownChan)
			s.shutdownChan <- s.Shutdown(sCtx)
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-startChan:
			if err != nil {
				return err
			}
		case <-time.After(5 * time.Second):
			return errors.New("timeout starting server listeners")
		}
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.Shutdown(ctx)
}

func (s *Server) Wait() error {
	// TODO: think about our timeouts
	return <-s.shutdownChan
}

type NewServerParams struct {
	Mux       ServeMux
	Handlers  []*Handler
	Listeners []net.Listener
}

type ListenAddr string

func NewTCPListener(addr ListenAddr) (net.Listener, error) {
	return net.Listen("tcp", string(addr))
}

func NewServer(params *NewServerParams) (*Server, error) {
	for _, handler := range params.Handlers {
		params.Mux.Handle(handler)
	}

	return &Server{
		Server: &http.Server{
			Handler: params.Mux,
		},
		Listeners: params.Listeners,

		// buffered to avoid blocking on shutdown
		shutdownChan:    make(chan error, 1),
		shutdownTimeout: lo.ToPtr(time.Second * 30),
	}, nil
}
