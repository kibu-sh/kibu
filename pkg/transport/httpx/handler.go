package httpx

import (
	"github.com/discernhq/devx/pkg/transport"
	"net/http"
)

type RawHandlerFunc func(w http.ResponseWriter, r *http.Request) error

type Handler[Req, Res any] struct {
	Path            string
	Methods         []string
	Decoder         DecoderFunc[Req]
	Encode          EncoderFunc[Res]
	Endpoint        transport.Endpoint[Req, Res]
	PreRequestHooks []RawHandlerFunc
}

func (h *Handler[Req, Res]) Route() Route {
	return Route{
		Path:    h.Path,
		Methods: h.Methods,
	}
}

func NewHandler[Req, Res any](
	path string,
	endpoint transport.Endpoint[Req, Res],
) *Handler[Req, Res] {
	h := &Handler[Req, Res]{
		Path:     path,
		Endpoint: endpoint,
		Decoder:  Decode[Req],
		Encode:   Encode[Res],
		Methods:  []string{http.MethodGet},
	}

	return h
}

func (h *Handler[Req, Res]) WithMethods(methods ...string) *Handler[Req, Res] {
	h.Methods = methods
	return h
}

func (h *Handler[Req, Res]) WithPreRequestHooks(hooks ...RawHandlerFunc) *Handler[Req, Res] {
	h.PreRequestHooks = hooks
	return h
}

func (h *Handler[Req, Res]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, hook := range h.PreRequestHooks {
		if err := hook(w, r); err != nil {
			// TODO: handle error
			return
		}
	}

	var req Req

	// TODO: decode errors
	_ = h.Decoder(r)(r.Context(), &req)
	// encoder := h.Encode(w)

	res, err := h.Endpoint(r.Context(), req)
	if err != nil {
		return
	}

	// TODO: encode errors
	_ = h.Encode(w)(r.Context(), &res)
	return
}
