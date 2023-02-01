package foreman

import (
	"context"
	"github.com/discernhq/devx/pkg/utils"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
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
}

func NewManager(ctx context.Context) *Manager {
	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)
	return &Manager{
		errGroup: eg,
		ctx:      ctx,
		cancel:   cancel,
		tasks:    utils.NewSyncMap[Process](),
	}
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

	m.tasks.Store(p.Name, &p)
	ready := make(chan struct{})

	m.errGroup.Go(func() error {
		return p.Start(m.ctx, func() {
			close(ready)
		})
	})

	select {
	case <-ready:
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
