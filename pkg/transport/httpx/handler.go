package httpx

import (
	"github.com/discernhq/devx/pkg/transport"
	"net/http"
)

// check if Handler implements http.Handler
var _ http.Handler = (*Handler)(nil)

// Handler is an HTTP adapter for transport.Handler
type Handler struct {
	Path    string
	Methods []string
	OnError func(err error)
	Handler transport.Handler
}

type HandlerOption func(h *Handler)

func NewHandler(path string, handler transport.Handler, opt ...HandlerOption) *Handler {
	return &Handler{
		Path:    path,
		Handler: handler,
		Methods: []string{http.MethodGet},
	}
}

func WithMethods(methods ...string) HandlerOption {
	return func(h *Handler) {
		h.Methods = methods
	}
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &Request{r}
	res := &ResponseWriter{w}
	if err := h.Handler.Serve(req.Context(), req, res); err != nil && h.OnError != nil {
		h.OnError(err)
	}
}

var _ = http.Handle
