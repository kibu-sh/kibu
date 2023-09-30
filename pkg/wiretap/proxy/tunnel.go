package proxy

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const (
	connectionEstablishedMessage = "HTTP/1.0 200 Connection established\r\n\r\n"
)

type TunnelMiddleware struct {
	mitmAddress   string
	logger        *slog.Logger
	dialTimeout   time.Duration
	tunnelTimeout time.Duration
}

// NewTunnelMiddleware creates a new tunnel middleware that will forward CONNECT requests to the MITM address.
// This is intended for use with testing and debugging.
func NewTunnelMiddleware(mitmAddress string) *TunnelMiddleware {
	return &TunnelMiddleware{
		mitmAddress: mitmAddress,
		logger:      slog.Default(),
		dialTimeout: time.Second * 5,
	}
}

func (t *TunnelMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if possiblyAProxyLoop(t.mitmAddress, r) {
			http.Error(w,
				fmt.Sprintf("loop detected for host %s", r.Host),
				http.StatusLoopDetected,
			)
			return
		}

		if isNotConnectRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		var err error
		var clientConn io.ReadWriteCloser

		ctx := r.Context()
		log := t.logger.With("host", r.Host)
		log.InfoContext(ctx, "handling connect tunnel request")

		defer recoverFromTunnelPanic(ctx, t.logger)
		//defer logErrorAndCloseClientConnection(ctx, &err, t.logger, clientConn)

		controller := http.NewResponseController(w)
		clientConn, _, err = controller.Hijack()
		if err != nil {
			return
		}

		mitmConn, err := net.DialTimeout("tcp", t.mitmAddress, t.dialTimeout)
		if err != nil {
			return
		}

		log.InfoContext(ctx, "tunnel established")
		_, err = clientConn.Write([]byte(connectionEstablishedMessage))
		if err != nil {
			return
		}

		go copyAndClose(ctx, log, mitmConn, clientConn, "mitm -> client")
		go copyAndClose(ctx, log, clientConn, mitmConn, "mitm <- client")
	})
}

func possiblyAProxyLoop(mitmAddress string, req *http.Request) bool {
	_, mitmPort, _ := net.SplitHostPort(mitmAddress)
	targetHost, targetPort, _ := net.SplitHostPort(req.Host)
	return isConnectRequest(req) &&
		isLoopBack(targetHost) &&
		portsMatch(targetPort, mitmPort)
}

func portsMatch(targetPort string, mitmPort string) bool {
	return targetPort == mitmPort
}

func isLoopBack(host string) bool {
	return host == "localhost" || host == "127.0.0.1"
}

func isConnectRequest(r *http.Request) bool {
	return r.Method == http.MethodConnect
}

func isNotConnectRequest(r *http.Request) bool {
	return !isConnectRequest(r)
}

func recoverFromTunnelPanic(ctx context.Context, logger *slog.Logger) {
	if r := recover(); r != nil {
		logger.ErrorContext(ctx, "panic in tunnel middleware", "panic", r)
	}
}

func copyAndClose(
	ctx context.Context,
	logger *slog.Logger,
	src io.ReadCloser,
	dest io.WriteCloser,
	direction string,
) {
	defer func(src io.ReadCloser) {
		_ = src.Close()
	}(src)

	defer func(dest io.WriteCloser) {
		_ = dest.Close()
	}(dest)

	_, err := io.Copy(dest, src)

	if err != nil {
		logger.DebugContext(ctx, "error reading data",
			"direction", direction,
			"error", err,
		)
	}
}
