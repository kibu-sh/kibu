package httpx

import (
	"github.com/discernhq/devx/pkg/transport"
	"net/http"
)

type Handler[Req, Res any] struct {
	Path     string
	Methods  []string
	Decoder  DecoderFunc[Req]
	Encode   EncoderFunc[Res]
	Endpoint transport.Endpoint[Req, Res]
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
	methods ...string,
) *Handler[Req, Res] {
	if len(methods) == 0 {
		methods = []string{http.MethodGet}
	}

	h := &Handler[Req, Res]{
		Path:     path,
		Methods:  methods,
		Endpoint: endpoint,
		Decoder:  Decode[Req],
		Encode:   Encode[Res],
	}

	return h
}

func (h *Handler[Req, Res]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
