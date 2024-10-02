package httpx

import (
	"fmt"
	"github.com/kibu-sh/kibu/pkg/slogx"
	"github.com/kibu-sh/kibu/pkg/transport"
	"github.com/kibu-sh/kibu/pkg/transport/middleware"
	"github.com/pkg/errors"
	"log/slog"
	"net/http"
	"time"
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
	ctx := r.Context()
	logger := slog.Default()
	tctx := &Context{
		req:    req,
		writer: res,
		codec:  h.Codec,
	}

	startTime := time.Now()
	var serveError error
	var encodingError error

	defer func() {
		stopTime := time.Now()
		duration := stopTime.Sub(startTime)
		logger = slogx.BindLogBuilders(logger,
			slogx.WithRequestInfo(req),
			slogx.WithResponseInfo(res),
			slogx.WithDuration(duration),
			slogx.WithErrorInfo(serveError),
		)

		level := slog.LevelInfo
		if serveError != nil || encodingError != nil {
			level = slog.LevelError
		}

		if encodingError != nil {
			logger = logger.With("encoding.error", encodingError)
		}

		logger.Log(ctx, level, buildHTTPResponseLogMessage(req, res))
	}()

	// if there's no error from the serve handler, it means the request was successful,
	// there's no need to encode an error to the transport
	if serveError = h.Handler.Serve(tctx); serveError == nil {
		return
	}

	if encodingError = tctx.Codec().EncodeError(ctx, res, serveError); encodingError != nil {
		encodingError = errors.Wrap(encodingError, "failed to write error to transport response")
	}
}

func buildHTTPResponseLogMessage(req transport.Request, res transport.Response) string {
	return fmt.Sprintf("%s %s %d %d %s",
		req.Version(),
		req.Method(),
		res.GetStatusCode(),
		res.BytesWritten(),
		req.URL().Path,
	)
}

var _ http.Handler = (*Handler)(nil)

type HandlerFactory interface {
	HTTPHandlerFactory(*middleware.Registry) []*Handler
}
type HandlerFactoryFunc func() []*Handler

func (h HandlerFactoryFunc) HTTPHandlerFactory() []*Handler {
	return h()
}
