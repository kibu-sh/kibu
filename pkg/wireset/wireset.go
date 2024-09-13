package wireset

import (
	"context"
	"fmt"
	"github.com/google/wire"
	"github.com/kibu-sh/kibu/pkg/appcontext"
	"github.com/kibu-sh/kibu/pkg/config"
	"github.com/kibu-sh/kibu/pkg/foreman"
	"github.com/kibu-sh/kibu/pkg/transport/httpx"
	"github.com/kibu-sh/kibu/pkg/transport/temporal"
	"github.com/kibu-sh/kibu/pkg/workspace"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"log/slog"
	"net"
	"net/http"
)

func ProvideServerAddress() httpx.ListenAddr {
	// 638725 spells neural on a phone keypad
	return httpx.ListenAddr(net.JoinHostPort("127.0.0.1", "6387"))
}

func NewListeners(
	ctx context.Context,
	addr httpx.ListenAddr,
	store config.Store,
) (listeners []net.Listener, err error) {
	tcpl, err := httpx.NewTCPListener(addr)
	if err != nil {
		return
	}
	listeners = append(listeners, tcpl)
	return
}

func NewTemporalOptions(ctx context.Context, store config.Store) (opts client.Options, err error) {
	_, err = store.GetByKey(ctx, "temporal", &opts)
	return
}

func NewTemporalClient(
	opts client.Options,
	log *slog.Logger,
) (c client.Client, err error) {
	opts.Logger = log

	c, err = client.Dial(opts)
	if err != nil {
		err = errors.Wrap(err, "failed to connect to temporal")
		return
	}
	return
}

func NewConfigStore() (store config.Store, err error) {
	store, err = workspace.DefaultConfigStore("dev")
	return
}

func NewLogger() *slog.Logger {
	return slog.Default()
}

func NewForeman(
	ctx context.Context,
	server *http.Server,
	listeners []net.Listener,
	workers []worker.Worker,
	logger *slog.Logger,
) (m *foreman.Manager, err error) {
	m = foreman.NewManager(ctx, foreman.WithLogger(logger))
	for i, wrk := range workers {
		err = m.Register(foreman.NewProcess(
			fmt.Sprintf("temporal-worker-%d", i), startWorker(wrk, logger),
		))
		if err != nil {
			return
		}
	}

	for _, listener := range listeners {
		name := fmt.Sprintf("server %s", listener.Addr().String())
		err = m.Register(foreman.NewProcess(name, startListener(listener, server, logger)))
		if err != nil {
			return
		}
	}

	return
}

func startListener(
	listener net.Listener,
	server *http.Server,
	logger *slog.Logger,
) func(ctx context.Context, ready func()) error {
	return func(ctx context.Context, ready func()) (err error) {
		if err = httpx.StartServer(ctx, listener, server); err != nil {
			return err
		}
		ready()
		<-ctx.Done()

		logger.Debug("shutting down http server")
		err = server.Shutdown(ctx)
		return
	}
}

func startWorker(wrk worker.Worker, logger *slog.Logger) foreman.StartFunc {
	return func(ctx context.Context, ready func()) (err error) {
		if err = wrk.Start(); err != nil {
			return
		}
		ready()
		logger.Debug("temporal worker ready")
		<-ctx.Done()

		logger.Debug("shutting down temporal worker")
		wrk.Stop()
		return
	}
}

var DefaultSet = wire.NewSet(
	NewForeman,
	ProvideServerAddress,
	NewListeners,
	NewTemporalClient,
	NewTemporalOptions,
	NewConfigStore,
	NewLogger,

	appcontext.Context,
	temporal.NewWorker,
	httpx.NewServer,
	httpx.NewTCPListener,
	httpx.NewStdLibMux,

	wire.Bind(new(httpx.ServeMux), new(*httpx.StdLibMux)),
	wire.Struct(new(httpx.NewServerParams), "*"),
	wire.Value(temporal.TaskQueue("default")),
)
