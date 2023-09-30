package proxy

import (
	"context"
	"crypto/tls"
	"github.com/discernhq/devx/pkg/wiretap/certgen"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type TunnelMiddlewareSuite struct {
	suite.Suite
	connectListener net.Listener
	connectServer   *http.Server

	mitmListener net.Listener
	mitmServer   *http.Server

	certPool   *certgen.DynamicCertPool
	httpClient *http.Client
	url        *url.URL
}

func TestHTTPTunnelListenerSuite(t *testing.T) {
	suite.Run(t, new(TunnelMiddlewareSuite))
}

func (s *TunnelMiddlewareSuite) SetupSuite() {
	var err error
	r := s.Require()
	s.certPool = certgen.NewDefaultCertPool()
	s.connectListener, err = net.Listen("tcp", "localhost:0")
	r.NoError(err)

	s.connectListener = tls.NewListener(s.connectListener, &tls.Config{
		GetCertificate: s.certPool.GetCertificateByHello,
	})

	s.mitmListener, err = net.Listen("tcp", "localhost:0")
	r.NoError(err)

	s.mitmListener = tls.NewListener(s.mitmListener, &tls.Config{
		GetCertificate: s.certPool.GetCertificateByHello,
	})

	tunnel := NewTunnelMiddleware(s.mitmListener.Addr().String())

	s.connectServer = &http.Server{
		Handler: tunnel.Handler(nil),
	}

	s.mitmServer = &http.Server{
		Handler: tunnel.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-intercepted", "true")
			_, _ = w.Write([]byte("hello"))
		})),
	}

	s.url, err = url.Parse("https://" + s.connectListener.Addr().String())
	r.NoError(err)
	s.httpClient = NewClient(s.url, s.certPool.ToTLSConfig())
}

func (s *TunnelMiddlewareSuite) TearDownSuite() {
	_ = s.connectListener.Close()
	_ = s.mitmListener.Close()
}

func (s *TunnelMiddlewareSuite) TestHTTPTunnelListener() {
	r := s.Require()
	go func() {
		_ = s.mitmServer.Serve(s.mitmListener)
	}()

	go func() {
		_ = s.connectServer.Serve(s.connectListener)
	}()

	res, err := s.httpClient.Get("https://postman-echo.com/get")
	r.NoError(err)
	r.Equal(200, res.StatusCode)
	r.Equal("true", res.Header.Get("X-intercepted"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, "https://"+s.mitmListener.Addr().String(), nil)
	r.NoError(err)

	req = req.WithContext(ctx)

	res, err = s.httpClient.Do(req)
	r.NotNil(err, "expected error for loop detection")
	r.Falsef(errors.Is(err, context.DeadlineExceeded), "should not have deadline exceeded error")
	r.Contains(err.Error(), "Loop Detected", "server should prevent proxy loops")
}
