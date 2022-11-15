package httpx

import (
	"context"
	"go.uber.org/fx"
	"net"
	"net/http"
)

type Server interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type NetListener struct {
	*http.Server
	l net.Listener
}

func (s *NetListener) Start(ctx context.Context) error {
	return s.Serve(s.l)
}

func (s *NetListener) Stop(ctx context.Context) error {
	return s.Shutdown(ctx)
}

func OnStart(ctx context.Context, s Server) error {
	go s.Start(ctx)
	return nil
}

func OnStop(ctx context.Context, s Server) error {
	return s.Stop(ctx)
}

type NewServerParams struct {
	Addr string
	Mux  *ServeMux
}

func NewNetListener(params *NewServerParams) (Server, error) {
	s := &http.Server{
		Addr:    params.Addr,
		Handler: params.Mux,
	}

	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return nil, err
	}

	return &NetListener{
		s, l,
	}, nil
}

func AsController(controller any) fx.Option {
	return fx.Provide(
		fx.Annotate(
			controller,
			fx.As(new(Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	)
}
