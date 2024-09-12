package adminserver

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/spec"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"net/url"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"testing"
)

type AdminServerSuite struct {
	suite.Suite
	server  *httptest.Server
	httpURL *url.URL
	wsURL   *url.URL
	topic   spec.SnapshotMessageTopic
}

func TestAdminServerSuite(t *testing.T) {
	suite.Run(t, new(AdminServerSuite))
}

func (s *AdminServerSuite) SetupSuite() {
	var topic = spec.NewSnapshotMessageTopic()
	mux := http.NewServeMux()
	mux = BindToMux(mux, topic)
	s.topic = topic
	s.server = httptest.NewServer(mux)
	s.httpURL, _ = url.Parse(s.server.URL)
	s.wsURL = toWebSocketURL(s.httpURL)
}

func (s *AdminServerSuite) TearDownSuite() {
	s.server.Close()
}

func (s *AdminServerSuite) TestAdminSnapshotStream() {
	ctx := context.Background()
	r := s.Require()
	expected := spec.Snapshot{
		ID: "test",
	}

	snapshotURL := s.wsURL.JoinPath(Endpoints.SnapshotStream).String()
	wsClient, _, err := websocket.Dial(ctx, snapshotURL, nil)
	r.NoError(err)

	err = s.topic.Publish(ctx, expected)
	r.NoError(err)

	var snapshot spec.Snapshot
	err = wsjson.Read(ctx, wsClient, &snapshot)
	r.NoError(err)
	r.EqualValues(expected, snapshot)
}

func toWebSocketURL(u *url.URL) *url.URL {
	u2 := *u
	u2.Scheme = "ws"
	return &u2
}
