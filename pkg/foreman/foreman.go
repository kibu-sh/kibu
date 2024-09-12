package foreman

import (
	"context"
	"fmt"
	"github.com/kibu-sh/kibu/pkg/utils"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"time"
)

type StartFunc func(ctx context.Context, ready func()) error

type Process struct {
	Name         string
	Start        StartFunc
	StartTimeout time.Duration
}

type Manager struct {
	tasks    *utils.SyncMap[Process]
	errGroup *errgroup.Group
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *slog.Logger
}

type Option func(m *Manager)

func DefaultOptions() []Option {
	return []Option{
		WithLogger(slog.Default()),
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(m *Manager) {
		m.logger = logger
	}
}

func NewManager(ctx context.Context, opts ...Option) *Manager {
	opts = append(DefaultOptions(), opts...)
	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)
	m := &Manager{
		errGroup: eg,
		ctx:      ctx,
		cancel:   cancel,
		tasks:    utils.NewSyncMap[Process](),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

var ErrProcessExists = errors.New("process already exists")
var ErrProcessFailedToReadyUp = errors.New("process failed to ready up")

func (m *Manager) Shutdown() {
	m.cancel()
}

func (m *Manager) Register(p Process) (err error) {
	if m.tasks.Has(p.Name) {
		err = errors.Wrap(ErrProcessExists, p.Name)
		return
	}

	log := m.logger.With("proc", p.Name)

	log.Debug(fmt.Sprintf("[kibu.foreman] registering process: %s", p.Name))

	m.tasks.Store(p.Name, &p)
	ready := make(chan struct{})

	m.errGroup.Go(func() error {
		return p.Start(m.ctx, func() {
			close(ready)
		})
	})

	select {
	case <-ready:
		log.Info(fmt.Sprintf("[kibu.foreman] %s ready", p.Name))
		return nil
	case <-m.ctx.Done():
		err = errors.Wrapf(m.ctx.Err(), "failed to start proc: %s", p.Name)
		return
	case <-time.After(p.StartTimeout):
		err = errors.Wrapf(ErrProcessFailedToReadyUp, "proc: %s", p.Name)
		return
	}
}

func (m *Manager) Wait() error {
	return m.errGroup.Wait()
}

func NewProcess(name string, start StartFunc) Process {
	return Process{
		Name:         name,
		Start:        start,
		StartTimeout: 5 * time.Second,
	}
}
