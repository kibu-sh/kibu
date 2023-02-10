package httpx

import (
	"github.com/discernhq/devx/pkg/transport"
	"github.com/discernhq/devx/pkg/transport/middleware"
	"net/http"
)

// check if Handler implements http.Handler
var _ http.Handler = (*Handler)(nil)

// Handler is an HTTP adapter for transport.Handler
type Handler struct {
	Path    string
	Methods []string
	Handler transport.Handler
	Codec   transport.Codec

	// TODO: think about emitting errors at a higher level
	// Maybe we need a logger here
	OnError func(err error)
}

func NewHandler(path string, handler transport.Handler) *Handler {
	return &Handler{
		Path:    path,
		Handler: handler,
		Methods: []string{http.MethodGet},
		Codec:   DefaultCodec,
	}
}

func (h *Handler) WithMethods(methods ...string) *Handler {
	h.Methods = methods
	return h
}

// TODO: consider capturing panics at this level

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := NewRequest(r)
	res := NewResponse(w)
	ctx := &Context{
		Context: r.Context(),
		req:     req,
		writer:  res,
		codec:   h.Codec,
	}

	if err := h.Handler.Serve(ctx); err != nil {
		_ = ctx.Codec().EncodeError(ctx, ctx.Response(), err)
	}
}

var _ http.Handler = (*Handler)(nil)

type HandlerFactory interface {
	HTTPHandlerFactory(*middleware.Registry) []*Handler
}
type HandlerFactoryFunc func() []*Handler

func (h HandlerFactoryFunc) HTTPHandlerFactory() []*Handler {
	return h()
}
