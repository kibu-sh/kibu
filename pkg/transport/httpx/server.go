package httpx

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type NewServerParams struct {
	Mux      ServeMux
	Handlers []*Handler
}

type ListenAddr string

func NewTCPListener(addr ListenAddr) (net.Listener, error) {
	return net.Listen("tcp", string(addr))
}

func NewServer(params *NewServerParams) (*http.Server, error) {
	log := slog.Default()
	for _, handler := range params.Handlers {
		log.Debug(fmt.Sprintf("[kibue.transport.httpx] %s %s",
			handler.Methods, handler.Path))
		params.Mux.Handle(handler)
	}

	return &http.Server{
		Handler: params.Mux,
	}, nil
}

var ErrServerStartTimeout = errors.New("http server start timeout")

// StartServer starts the server in a non-blocking fashion
// It will return an error in the following cases
// 1. The server fails to start (port conflict)
// 2. The context is cancelled
// 3. The server takes longer than 5 seconds to start
// You must call server.Shutdown() to stop the server
func StartServer(
	ctx context.Context,
	listener net.Listener,
	server *http.Server,
) error {
	ready := make(chan struct{})
	errCh := make(chan error)
	go func() {
		close(ready)
		slog.Default().
			With("address", listener.Addr().String()).
			Info(fmt.Sprintf("starting server on %s", listener.Addr().String()))
		errCh <- server.Serve(listener)
	}()

	select {
	case <-ready:
		return nil
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return ErrServerStartTimeout
	}
}
