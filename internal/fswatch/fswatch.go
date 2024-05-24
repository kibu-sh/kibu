package fswatch

import (
	"context"
	"errors"
	"github.com/fsnotify/fsnotify"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type Event struct {
	fsnotify.Event

	// GitOp indicates if the event occurred while a git operation was in progress
	// This will be true when GitRoot/.git/index.lock exists
	GitOp bool
}

var (
	ErrReadTimeout = errors.New("failed to read an event within the timeout")
	ErrStopTimeout = errors.New("recursive watcher did not shutdown within the timeout")
)

type RecursiveWatcher struct {
	dir       string
	gitRoot   string
	ctx       context.Context
	cancelCtx context.CancelFunc
	watcher   *fsnotify.Watcher
	events    chan Event
	stop      chan error
	errors    chan error
}

type Option func(*RecursiveWatcher) error

func WithGitRoot(gitRoot string) Option {
	return func(rw *RecursiveWatcher) error {
		rw.gitRoot = gitRoot
		return nil
	}
}

func WithEventsChannel(events chan Event) Option {
	return func(rw *RecursiveWatcher) error {
		rw.events = events
		return nil
	}
}

func gitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func WithDetectGitRoot() Option {
	return func(rw *RecursiveWatcher) error {
		root, err := gitRoot()
		if err != nil {
			return err
		}

		return WithGitRoot(root)(rw)
	}
}

func NewRecursiveWatcher(ctx context.Context, dir string, opts ...Option) (rw *RecursiveWatcher, err error) {
	rw = &RecursiveWatcher{
		dir:    dir,
		events: make(chan Event),
		stop:   make(chan error, 1),
	}

	rw.ctx, rw.cancelCtx = context.WithCancel(ctx)

	for _, opt := range opts {
		if err = opt(rw); err != nil {
			return
		}
	}

	rw.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return
	}

	if err = rw.AddDir(rw.dir); err != nil {
		return
	}

	return
}

func (rw *RecursiveWatcher) Events() <-chan Event {
	return rw.events
}

func (rw *RecursiveWatcher) Errors() <-chan error {
	return rw.errors
}

func (rw *RecursiveWatcher) ReadOne(timeout time.Duration) (Event, error) {
	select {
	case e := <-rw.events:
		return e, nil
	case <-time.After(timeout):
		return Event{}, ErrReadTimeout
	}
}

var ignoredDirs = []string{
	".git",
	".idea",
	".vscode",
	"node_modules",
	"vendor",
}

func isIgnoredDir(d fs.DirEntry) bool {
	return slices.Contains(ignoredDirs, filepath.Base(d.Name()))
}

func (rw *RecursiveWatcher) AddDir(path string) error {
	if err := rw.watcher.Add(path); err != nil {
		return err
	}
	return fs.WalkDir(os.DirFS(path), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && isIgnoredDir(d) {
			return fs.SkipDir
		}

		if d.IsDir() {
			return rw.watcher.Add(path)
		}
		return nil
	})
}

// Stop stops the RecursiveWatcher and waits for it to shut down.
// If the watcher does not shut down within the specified timeout,
// an error is returned.
func (rw *RecursiveWatcher) Stop(timeout time.Duration) (err error) {
	rw.cancelCtx()
	select {
	case err = <-rw.stop:
	case <-time.After(timeout):
		err = ErrStopTimeout
	}
	return
}

// Start starts the RecursiveWatcher by running the event loop in a separate goroutine.
// This method waits until the goroutine has started before returning.
// Call Stop to stop the watcher and wait for it to shut down.
func (rw *RecursiveWatcher) Start() {
	started := make(chan struct{})
	go func() {
		close(started)
		defer close(rw.stop)
		err := rw.runEventLoop()
		rw.stop <- err
	}()
	<-started
}

// runEventLoop is a method that continuously listens for events from the file watcher and processes them.
// It runs in an infinite loop until the context is done, or the watcher channel is closed.
// If an error occurs or the context is done, the loop is exited and cleanup is performed.
func (rw *RecursiveWatcher) runEventLoop() (err error) {
	for {
		select {
		case <-rw.ctx.Done():
			err = rw.ctx.Err()
			goto cleanup
		case ev, ok := <-rw.watcher.Events:
			if !ok {
				goto cleanup
			}
			rw.processEvent(ev)
		case err, ok := <-rw.watcher.Errors:
			if !ok {
				goto cleanup
			}
			rw.processError(err)
		}
	}

cleanup:
	close(rw.events)
	return errors.Join(err, rw.watcher.Close())
}

// processEvent processes a fsnotify.Event and sends a mapped Event through the events channel.
// It checks if the event occurred while a git operation was in progress by checking for the presence of .git/index.lock.
// The mapped Event contains the name and operation of the original event, as well as a op indicating if it was a git operation.
// The mapped Event is sent through the events channel for further processing.
func (rw *RecursiveWatcher) processEvent(event fsnotify.Event) {
	// chmod events should be ignored
	// they are triggered frequently by external software like antivirus and macOS spotlight
	if event.Has(fsnotify.Chmod) {
		return
	}

	// Check if the event occurred while a git operation was in progress
	rw.events <- Event{
		Event: event,
		// git operations are indicated by the presence of .git/index.lock
		GitOp: fileExists(filepath.Join(rw.gitRoot, ".git", "index.lock")),
	}

	// bail early if the event is not a directory creation event
	if !event.Has(fsnotify.Create) && !isDir(event.Name) {
		return
	}

	rw.processError(rw.AddDir(event.Name))
}

func (rw *RecursiveWatcher) processError(err error) {
	select {
	case rw.errors <- err:
	default:
		// drop errors if the errors channel is full, or uninitialized
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func DebounceOnGitOp(e Event) bool {
	return e.GitOp
}

func IsGeneratedFile(e Event) bool {
	return strings.Contains(e.Name, "gen")
}
