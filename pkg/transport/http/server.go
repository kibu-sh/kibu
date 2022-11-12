package http

import "net/http"

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

type Server struct {
	Router      Router
	Controllers []Controller
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

func NewServer(router Router, controllers []Controller) *Server {
	for _, controller := range controllers {
		router.MountController(controller)
	}
	return &Server{
		Router:      router,
		Controllers: controllers,
	}
}
