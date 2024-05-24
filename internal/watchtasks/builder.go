package watchtasks

import (
	"context"
	"github.com/discernhq/devx/internal/fswatch"
	"github.com/discernhq/devx/pkg/messaging"
	"github.com/discernhq/devx/pkg/messaging/multichannel"
	"github.com/pkg/errors"
	"log/slog"
	"os"
	"time"
)

var ErrStopTimeout = errors.New("failed to stop builder within timeout")

type Command struct {
	Cmd  string
	Args []string
}

// Builder watches buffered fs events and executes a build task
type Builder struct {
	rootCtx  context.Context
	shutdown context.CancelFunc
	tmpDir   string
	rootDir  string
	appDir   string
	fsEvents chan []fswatch.Event
	topic    *multichannel.Topic[Event]
	log      *slog.Logger

	finished          chan error
	rebuild           chan struct{}
	restart           chan struct{}
	debuggerListening chan struct{}
}

type BuilderOption func(*Builder) error

func WithRootDir(dir string) BuilderOption {
	return func(b *Builder) error {
		b.rootDir = dir
		return nil
	}
}

func WithAppDir(dir string) BuilderOption {
	return func(b *Builder) error {
		b.appDir = dir
		return nil
	}
}

func WithFSEvents(events chan []fswatch.Event) BuilderOption {
	return func(b *Builder) error {
		b.fsEvents = events
		return nil
	}
}

func (b *Builder) Start() {
	started := make(chan struct{})
	go func() {
		close(started)
		b.runEventLoop()
	}()
	<-started
	return
}

func (b *Builder) Stop(timeout time.Duration) (err error) {
	b.shutdown()
	select {
	case err = <-b.finished:
	case <-time.After(timeout):
		err = ErrStopTimeout
	}
	return
}

func (b *Builder) Subscribe(ctx context.Context) (messaging.Stream[Event], error) {
	return b.topic.Subscribe(ctx)
}

func NewBuilder(ctx context.Context, opts ...BuilderOption) (b *Builder, err error) {
	b = &Builder{
		finished:          make(chan error),
		fsEvents:          make(chan []fswatch.Event),
		rebuild:           make(chan struct{}),
		restart:           make(chan struct{}),
		debuggerListening: make(chan struct{}),
		topic:             multichannel.NewTopicWithDefaults[Event](),
		log:               slog.Default(),
	}

	b.rootCtx, b.shutdown = context.WithCancel(ctx)
	b.tmpDir, err = os.MkdirTemp("", "devx-buildout-*")
	if err != nil {
		return
	}

	for _, opt := range opts {
		if err = opt(b); err != nil {
			return
		}
	}

	return b, nil
}
