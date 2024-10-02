package spec

import (
	"bytes"
	"compress/gzip"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var _ suite.SetupAllSuite = (*GzipSuite)(nil)
var _ suite.TearDownAllSuite = (*GzipSuite)(nil)

type GzipSuite struct {
	suite.Suite
	server *httptest.Server
}

func TestGzipSuite(t *testing.T) {
	suite.Run(t, new(GzipSuite))
}

func (s *GzipSuite) SetupSuite() {
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var res io.Writer = w
		var gzres = gzip.NewWriter(w)
		defer gzres.Close()
		if strings.EqualFold(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			res = gzres
		}
		_, _ = res.Write([]byte("hello"))
	})
	s.server = httptest.NewServer(handler)
}

func (s *GzipSuite) TearDownSuite() {
	s.server.Close()
}

func (s *GzipSuite) TestGzipDecompression() {
	r := s.Require()
	s.server.Client()
	req, err := http.NewRequest(http.MethodGet, s.server.URL, nil)
	r.NoError(err)
	//req.Header.Set("Accept-Encoding", "gzip")

	res, err := s.server.Client().Do(req)
	r.NoError(err)
	r.Truef(res.Uncompressed, "response should be uncompressed")

	r.Equal(http.StatusOK, res.StatusCode)
	r.Equal("", res.Header.Get("Content-Encoding"))

	var body bytes.Buffer
	_, err = body.ReadFrom(res.Body)
	r.NoError(err)

	r.Equal("hello", string(body.Bytes()))
}
