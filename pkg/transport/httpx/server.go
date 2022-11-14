package httpx

import (
	"context"
	"go.uber.org/fx"
	"net"
	"net/http"
)

type Route struct {
	Path    string
	Methods []string
}

type Controller interface {
	http.Handler
	Route() Route
}

type Router interface {
	http.Handler
	MountController(controller Controller)
}

type StdRouter struct {
	*http.ServeMux
}

func NewStdRouter() Router {
	return StdRouter{http.NewServeMux()}
}

func (s StdRouter) MountController(controller Controller) {
	route := controller.Route()
	s.Handle(route.Path, controller)
}

type ServeMux struct {
	Router Router
}

func (s ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

type NewServeMuxParams struct {
	fx.In

	Router      Router
	Controllers []Controller `group:"controllers"`
}

func NewServeMux(params NewServeMuxParams) *ServeMux {
	for _, controller := range params.Controllers {
		params.Router.MountController(controller)
	}
	return &ServeMux{
		Router: params.Router,
	}
}

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
	return s.Start(ctx)
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

func AsController(c any) fx.Option {
	return fx.Provide(
		fx.Annotate(c,
			fx.As(new(Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	)
}
