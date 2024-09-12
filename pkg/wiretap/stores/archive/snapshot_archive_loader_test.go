package archive

import (
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/internalmock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type LoadSnapshotsFromDirSuite struct {
	suite.Suite
	dir   string
	store SnapshotArchiveStore
}

func TestLoadSnapshotsFromDir(t *testing.T) {
	suite.Run(t, new(LoadSnapshotsFromDirSuite))
}

func (s *LoadSnapshotsFromDirSuite) SetupSuite() {
	s.dir = s.T().TempDir()
	s.store = NewSnapshotArchiveStore(s.dir)
}

func (s *LoadSnapshotsFromDirSuite) TestLoadSnapshotsFromDir() {
	sh := internalmock.NewTestSnapshot()
	ref1, err := s.store.Write(sh)
	s.Require().NoError(err)

	snapshots, err := LoadSnapshotsFromDir(s.dir)
	s.Require().NoError(err)
	s.Require().Len(snapshots, 1)
	s.Require().Equal(ref1.ID, snapshots[0].Ref().ID)
}
