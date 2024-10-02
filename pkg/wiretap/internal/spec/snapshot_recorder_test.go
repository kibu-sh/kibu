package spec

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"
)

var startTime = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

func TestSnapshotRecorderSuite(t *testing.T) {
	suite.Run(t, new(SnapshotRecorderSuite))
}

type SnapshotRecorderSuite struct {
	suite.Suite
	topic       SnapshotMessageTopic
	proxyServer *httptest.Server
	ctx         context.Context
	echoServer  *httptest.Server
}

func (s *SnapshotRecorderSuite) SetupSuite() {
	s.ctx = context.Background()
	s.topic = NewSnapshotMessageTopic()
	s.echoServer = httptest.NewServer(compressedEchoHandler())
	s.proxyServer = httptest.NewServer(NewSnapshotMiddleware(
		s.topic, deterministicTime(startTime))(
		httputil.NewSingleHostReverseProxy(lo.Must(url.Parse(s.echoServer.URL))),
	))
}

func (s *SnapshotRecorderSuite) TearDownSuite() {
	s.proxyServer.Close()
	s.echoServer.Close()
}

func (s *SnapshotRecorderSuite) TestRecordRequest() {
	r := s.Require()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	stream, err := s.topic.Subscribe(ctx)
	r.NoError(err)

	expected := "hello world"
	req, err := http.NewRequest(http.MethodPost, s.proxyServer.URL,
		io.NopCloser(strings.NewReader(expected)))
	r.NoError(err)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	res, err := s.proxyServer.Client().Do(req.WithContext(ctx))
	r.NoError(err)
	r.Equal(http.StatusOK, res.StatusCode)
	var body bytes.Buffer
	_, err = io.Copy(&body, res.Body)
	r.NoError(err)
	r.Equal(expected, string(body.Bytes()))

	sh, _, err := stream.Next(ctx)
	r.NoError(err)
	r.Equal(30*time.Second, sh.Duration)
	r.Equal(expected, sh.Request.Body)
	r.Equal(expected, sh.Response.Body)
	r.Equal(http.MethodPost, sh.Request.Method)
	r.Equal(http.StatusOK, sh.Response.StatusCode)
	r.Equal(http.StatusText(http.StatusOK), sh.Response.Status)
	r.Equal(s.proxyServer.URL+"/", sh.Request.URL.String())
	r.Equal("text/plain; charset=utf-8", sh.Request.Header.Get("Content-Type"))
	r.Equal("text/plain; charset=utf-8", sh.Response.Header.Get("Content-Type"))
	r.Equal(int64(11), sh.Request.ContentLength)
	r.Equal(int64(11), sh.Response.ContentLength)
}

func compressedEchoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		gzw := gzip.NewWriter(w)
		_, _ = gzw.Write(data)
		_ = gzw.Close()
	}
}

// deterministicTime returns a function that returns a time.Time
// the first call returns start
// duration is increased by 30s after each call
// each subsequent call returns start + duration
func deterministicTime(start time.Time) func() time.Time {
	duration := time.Second * 0
	return func() time.Time {
		now := start.Add(duration)
		duration += time.Second * 30
		return now
	}
}
