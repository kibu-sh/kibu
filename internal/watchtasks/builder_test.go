package watchtasks

import (
	"context"
	"cuelang.org/go/pkg/time"
	"github.com/discernhq/devx/internal/fswatch"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
)

func TestBuilderSuite(t *testing.T) {
	suite.Run(t, new(BuilderSuite))
}

type BuilderSuite struct {
	suite.Suite
	builder  *Builder
	fsEvents chan []fswatch.Event
	dir      string
}

func (s *BuilderSuite) SetupSuite() {
	var err error
	s.dir = s.T().TempDir()
	s.fsEvents = make(chan []fswatch.Event, 1)
	s.builder, err = NewBuilder(context.Background(),
		WithBuildSteps([]Command{
			{"touch", []string{
				filepath.Join(s.dir, "test.txt"),
			}},
		}))
	s.Require().NoError(err)
	s.builder.Start()
}

func (s *BuilderSuite) TearDownSuite() {
	_ = s.builder.Stop(time.Second * 3000)
}

func (s *BuilderSuite) TestBuildControl() {
	s.T().Run("should restart several times successfully", func(t *testing.T) {
		var count int

	loop:
		if count > 3 {
			return
		}
		count++
		s.builder.restart(nil)
		<-s.builder.stepsStarted
		err := <-s.builder.stepsFinished
		require.NoError(t, err)
		require.FileExists(t, filepath.Join(s.dir, "test.txt"))
		err = os.Remove(filepath.Join(s.dir, "test.txt"))
		require.NoError(t, err)
		goto loop
	})

	s.T().Run("should be able to cancel running commands", func(t *testing.T) {
		s.builder.steps = []Command{
			{"sleep", []string{"1000"}},
		}
		s.builder.restart(nil)
		<-s.builder.stepsStarted
		s.builder.cancelSteps()
		err := <-s.builder.stepsFinished
		require.ErrorIs(t, err, context.Canceled)
	})
}
