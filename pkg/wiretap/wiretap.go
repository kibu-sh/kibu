package wiretap

import (
	"bufio"
	"context"
	"github.com/discernhq/devx/pkg/messaging"
	"github.com/discernhq/devx/pkg/wiretap/certgen"
	"github.com/discernhq/devx/pkg/wiretap/internal/spec"
	"github.com/discernhq/devx/pkg/wiretap/proxy"
	"github.com/discernhq/devx/pkg/wiretap/servers/adminserver"
	"github.com/discernhq/devx/pkg/wiretap/servers/captureserver"
	"github.com/discernhq/devx/pkg/wiretap/servers/replayserver"
	"github.com/discernhq/devx/pkg/wiretap/stores/archive"
	"github.com/soheilhy/cmux"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	defaultHostPort = "localhost:9091"
)

type Server struct {
	addr        string
	logger      *slog.Logger
	store       spec.SnapshotStore
	topic       spec.SnapshotMessageTopic
	router      spec.SnapshotRouter
	certPool    spec.CertPool
	listener    net.Listener
	muxListener cmux.CMux
	// writeTopic is used to ensure that the capture writer
	// has a chance to write to disk before any other consumers may see the message
	writeTopic spec.SnapshotMessageTopic
}

func (s Server) WithAddr(addr string) Server {
	s.addr = addr
	return s
}

func (s Server) WithSnapshotDir(dir string) Server {
	s.store = archive.NewSnapshotArchiveStore(dir)
	return s
}

func (s Server) WithStore(store spec.SnapshotStore) Server {
	s.store = store
	return s
}

func (s Server) WithRouter(router spec.SnapshotRouter) Server {
	s.router = router
	return s
}

func (s Server) WithTopic(topic spec.SnapshotMessageTopic) Server {
	s.topic = topic
	return s
}

func (s Server) WithCertPool(pool spec.CertPool) Server {
	s.certPool = pool
	return s
}

func (s Server) StartInCaptureMode() (Server, error) {
	serveMux := http.NewServeMux()
	// mount admin endpoints
	serveMux = adminserver.BindToMux(serveMux, s.topic)
	// mount the capture proxy on the fallback route
	// bind to the write topic to ensure that the capture writer has a chance to write to disk
	// the snapshot will be published again over the main topic
	serveMux = captureserver.BindToMux(serveMux, s.writeTopic)
	// start the capture writer
	writeStream, err := s.writeTopic.Subscribe(context.Background())
	if err != nil {
		return s, err
	}

	// wait until the capture writer has started to start the server
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go s.runCaptureWriter(wg, writeStream)
	wg.Wait()
	return s.serve(serveMux)
}

func (s Server) StartInReplayMode() (Server, error) {
	serveMux := http.NewServeMux()
	// mount admin endpoints
	serveMux = adminserver.BindToMux(serveMux, s.topic)
	// mount the capture proxy on the fallback route
	serveMux = replayserver.BindToMux(serveMux, s.store, s.router, s.topic)
	return s.serve(serveMux)
}

func (s Server) runCaptureWriter(wg *sync.WaitGroup, writeStream messaging.Stream[spec.Snapshot]) {
	wg.Done()
	s.logger.Info("starting capture writer")
	defer writeStream.Unsubscribe()
	for snapshot := range writeStream.Channel() {
		if _, err := s.store.Write(&snapshot); err != nil {
			s.logger.Error("failed to write snapshot",
				"snapshot.id", snapshot.ID,
				"error", err)
		}

		if err := s.topic.Publish(context.Background(), snapshot); err != nil {
			s.logger.Error("failed to publish snapshot after write",
				"snapshot.id", snapshot.ID,
				"error", err)
		}
	}
}

func (s Server) serve(router *http.ServeMux) (Server, error) {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)

	if err != nil {
		return s, err
	}

	s.muxListener = cmux.New(s.listener)
	httpListener := s.muxListener.Match(matchInsecureHTTP)
	mitmListener := s.muxListener.Match(cmux.Any())

	tunnel := proxy.NewTunnelMiddleware(mitmListener.Addr().String())
	handler := tunnel.Handler(defaultLoggerMiddleware(router))

	connectServer := &http.Server{
		Handler: handler,
	}

	mitmServer := &http.Server{
		Handler:   handler,
		TLSConfig: s.certPool.ToTLSConfig(),
	}

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		wg.Done()
		_ = connectServer.Serve(httpListener)
	}()

	go func() {
		wg.Done()
		_ = mitmServer.ServeTLS(mitmListener, "", "")
	}()

	s.logger.Info("CONNECT proxy listening", "addr", s.URL().String())
	s.logger.Info("MITM proxy listening", "addr", s.SecureURL().String())

	wg.Wait()
	return s, nil
}

func (s Server) URL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   s.listener.Addr().String(),
	}
}

func (s Server) SecureURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   s.listener.Addr().String(),
	}
}

func (s Server) Serve() error {
	return s.muxListener.Serve()
}

func (s Server) WithLogger(logger *slog.Logger) Server {
	s.logger = logger
	return s
}

func (s Server) WithTestAddr() Server {
	s.addr = "localhost:0"
	return s
}

func (s Server) Close() {
	s.muxListener.Close()
	_ = s.listener.Close()
	return
}

func (s Server) Client() *http.Client {
	return proxy.NewClient(s.SecureURL(), s.certPool.ToTLSConfig())
}

func NewServer() *Server {
	return &Server{
		addr:       defaultHostPort,
		logger:     slog.Default(),
		writeTopic: spec.NewSnapshotMessageTopic(),
		topic:      spec.NewSnapshotMessageTopic(),
		certPool:   certgen.NewDefaultCertPool(),
	}
}

func defaultLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := slog.Default()
		start := time.Now()
		defer func() {
			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", time.Since(start).String(),
			)
		}()
		next.ServeHTTP(w, r)
	})
}

func matchInsecureHTTP(r io.Reader) bool {
	_, err := http.ReadRequest(bufio.NewReader(r))
	return err == nil
}
