package wiretap

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/messaging"
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/internalmock"
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/spec"
	"github.com/kibu-sh/kibu/pkg/wiretap/routers/dynamic"
	"github.com/kibu-sh/kibu/pkg/wiretap/rules/requestrules"
	"github.com/kibu-sh/kibu/pkg/wiretap/stores/archive"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type WiretapSuite struct {
	suite.Suite
	dir          string
	ctx          context.Context
	store        spec.SnapshotStore
	router       spec.SnapshotRouter
	echoServer   *httptest.Server
	topic        spec.SnapshotMessageTopic
	stream       messaging.Stream[spec.Snapshot]
	captureProxy Server
	replayProxy  Server
}

func TestWiretapSuite(t *testing.T) {
	suite.Run(t, new(WiretapSuite))
}

func (s *WiretapSuite) SetupSuite() {
	var err error
	s.ctx = context.Background()
	s.dir = s.T().TempDir()

	s.store = archive.NewSnapshotArchiveStore(s.dir)
	s.router = dynamic.NewSnapshotRouter()
	s.topic = spec.NewSnapshotMessageTopic()
	s.stream, _ = s.topic.Subscribe(s.ctx)

	s.captureProxy, err = NewServer().
		WithTestAddr().
		WithStore(s.store).
		WithTopic(s.topic).
		StartInCaptureMode()
	s.Require().NoError(err)

	s.replayProxy, err = NewServer().
		WithTestAddr().
		WithStore(s.store).
		WithTopic(s.topic).
		WithRouter(s.router).
		StartInReplayMode()
	s.Require().NoError(err)

	s.echoServer = httptest.NewServer(internalmock.EchoHandler())

	go s.replayProxy.Serve()
	go s.captureProxy.Serve()
}

func (s *WiretapSuite) TearDownSuite() {
	s.captureProxy.Close()
	s.replayProxy.Close()
	s.echoServer.Close()
}

func (s *WiretapSuite) TestProxyWithReplay() {
	var msg spec.Snapshot
	s.Run("should successfully proxy and capture a request", func() {
		r := s.Require()
		ctx, cancel := context.WithTimeout(s.ctx, time.Second*5)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.echoServer.URL, nil)
		r.NoError(err)
		req = req.WithContext(ctx)

		res, err := s.captureProxy.Client().Do(req)
		r.NoError(err)
		r.NotNil(res)
		r.Equal(http.StatusOK, res.StatusCode)

		msg, _, err := s.stream.Next(ctx)
		r.NoError(err)
		r.FileExists(archive.Filename(s.dir, msg.Ref()))

		snap, err := s.store.Read(msg.Ref())
		r.NoError(err)
		s.router.Register(msg.Ref(), requestrules.BasicMatchRule(snap))
	})

	s.Run("should successfully capture a round trip record", func() {
		r := s.Require()
		ctx, cancel := context.WithTimeout(s.ctx, time.Second*5)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.echoServer.URL, nil)
		r.NoError(err)
		req = req.WithContext(ctx)

		pRes, err := s.replayProxy.Client().Do(req)
		r.NoError(err)
		r.NotNil(pRes)

		snap, err := s.store.Read(msg.Ref())
		r.NoError(err)
		r.Equal(snap.Response.StatusCode, pRes.StatusCode)
	})

	s.Run("should fail to replay with no match", func() {
		r := s.Require()
		pRes, err := s.replayProxy.Client().Get(s.echoServer.URL + "/bad")
		r.NoError(err)
		r.NotNil(pRes)
		r.Equal(http.StatusBadGateway, pRes.StatusCode)
	})
}
