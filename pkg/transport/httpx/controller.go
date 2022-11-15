package httpx

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
