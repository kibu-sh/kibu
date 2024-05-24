package fswatch

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecursiveWatcherSuite(t *testing.T) {
	suite.Run(t, new(RecursiveWatcherSuite))
}

type RecursiveWatcherSuite struct {
	suite.Suite
	watcher *RecursiveWatcher
	dir     string
	root    fs.FS
}

func (s *RecursiveWatcherSuite) SetupSuite() {
	var err error
	r := s.Require()
	s.dir = s.T().TempDir()
	s.root = os.DirFS(s.dir)
	s.watcher, err = NewRecursiveWatcher(
		context.Background(), s.dir,
		WithGitRoot(s.dir),
	)
	r.NoError(err)
	s.watcher.Start()
}

func (s *RecursiveWatcherSuite) TearDownSuite() {
	err := s.watcher.Stop(time.Second * 5)
	s.Require().ErrorIs(err, context.Canceled)
}

func (s *RecursiveWatcherSuite) TestEventSequence() {
	type expect struct {
		op    fsnotify.Op
		name  string
		gitOp bool
	}

	type testCase struct {
		name     string
		action   func() error
		expected expect
	}

	tests := []testCase{
		{
			name: "should receive create event",
			action: func() error {
				return createDummyFile(s.dir, "test.txt")
			},
			expected: expect{
				op:   fsnotify.Create,
				name: filepath.Join(s.dir, "test.txt"),
			},
		},
		{
			name: "should receive write event",
			expected: expect{
				op:   fsnotify.Write,
				name: filepath.Join(s.dir, "test.txt"),
			},
			action: func() error {
				return os.WriteFile(filepath.Join(s.dir, "test.txt"), []byte("test"), 0644)
			},
		},
		{
			name: "should receive create event after renaming",
			expected: expect{
				op:   fsnotify.Create,
				name: filepath.Join(s.dir, "test2.txt"),
			},
			action: func() error {
				return os.Rename(filepath.Join(s.dir, "test.txt"), filepath.Join(s.dir, "test2.txt"))
			},
		},
		{
			name: "should receive a rename event after renaming",
			expected: expect{
				op:   fsnotify.Rename,
				name: filepath.Join(s.dir, "test.txt"),
			},
		},
		{
			name: "should receive remove event",
			expected: expect{
				op:   fsnotify.Remove,
				name: filepath.Join(s.dir, "test2.txt"),
			},
			action: func() error {
				return os.Remove(filepath.Join(s.dir, "test2.txt"))
			},
		},
		{
			name: "should receive create event after creating .git directory",
			expected: expect{
				op:   fsnotify.Create,
				name: filepath.Join(s.dir, ".git"),
			},
			action: func() error {
				return os.MkdirAll(filepath.Join(s.dir, ".git"), 0755)
			},
		},
		{
			name: "should show an operation as a GitOp while .git/index.lock is present",
			expected: expect{
				op:    fsnotify.Create,
				name:  filepath.Join(s.dir, ".git/index.lock"),
				gitOp: true,
			},
			action: func() error {
				return createDummyFile(s.dir, ".git/index.lock")
			},
		},
		{
			name: "should show create as a GitOp while .git/index.lock is present",
			expected: expect{
				op:    fsnotify.Create,
				name:  filepath.Join(s.dir, "dummy.txt"),
				gitOp: true,
			},
			action: func() error {
				return createDummyFile(s.dir, "dummy.txt")
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			if test.action != nil {
				require.NoError(t, test.action())
			}

			ev, err := s.watcher.ReadOne(time.Second * 3)
			require.NoError(t, err)
			require.Equal(t, test.expected.gitOp, ev.GitOp, "expected event to have desired git op")
			require.Equal(t, ev.Name, test.expected.name, "expected event to have desired name")
			require.Truef(t, ev.Has(test.expected.op), "expected event to have desired op")
		})
	}
}

func createDummyFile(dir, name string) error {
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	return f.Close()
}
