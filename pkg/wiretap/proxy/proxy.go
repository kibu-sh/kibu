package proxy

import (
	"context"
	"crypto/tls"
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/spec"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func New(tripper http.RoundTripper) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director:  newConnectProxyDirector(),
		Transport: tripper,
	}
}

// TODO: TEST THIS FUNCTION
// ENSURE WE PATCH REQUESTS COMING FROM A TUNNEL
func newConnectProxyDirector() func(req *http.Request) {
	return func(req *http.Request) {
		req.URL.Scheme = "http"
		if req.TLS != nil {
			req.URL.Scheme = "https"
		}

		req.URL.Host = req.Host
		req.RequestURI = req.URL.String()
		return
	}
}

func NewCaptureProxy(topic spec.SnapshotMessageTopic) http.Handler {
	return spec.NewSnapshotMiddleware(topic, time.Now)(New(http.DefaultTransport))
}

func NewReplayProxy(topic spec.SnapshotMessageTopic, replay *ReplayTransport) http.Handler {
	return spec.NewSnapshotMiddleware(topic, time.Now)(New(replay))
}

func NewTransport(proxyURL *url.URL, config *tls.Config) *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyURL(proxyURL),
		TLSClientConfig:       config,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: defaultTransportDialContext(&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}),
	}
}

func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}

func NewClient(proxyURL *url.URL, config *tls.Config) *http.Client {
	return &http.Client{
		Timeout:   time.Second * 60,
		Transport: NewTransport(proxyURL, config),
	}
}
