package httpx

import (
	"go.uber.org/fx"
	"net/http"
)

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
