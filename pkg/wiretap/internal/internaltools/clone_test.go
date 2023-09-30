package internaltools

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCloneRequestWithBody(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "https://localhost:8088/echo/sub/path?hello=echo",
		strings.NewReader(`{"hello": "echo"}`),
	)
	require.NoError(t, err)

	clone, err := CloneRequestWithBody(req)
	require.NoError(t, err)

	RequireReaderContentsMatch(t, req.Body, clone.Body)
	require.EqualValues(t, req.URL, clone.URL)
	require.EqualValues(t, req.Header, clone.Header)
	require.EqualValues(t, req.Method, clone.Method)
	require.EqualValues(t, req.Proto, clone.Proto)
	require.EqualValues(t, req.ProtoMajor, clone.ProtoMajor)
	require.EqualValues(t, req.ProtoMinor, clone.ProtoMinor)
	require.EqualValues(t, req.RequestURI, clone.RequestURI)
	require.EqualValues(t, req.ContentLength, clone.ContentLength)
	require.EqualValues(t, req.TransferEncoding, clone.TransferEncoding)
}

func TestCloneResponseWithBody(t *testing.T) {
	res := &http.Response{
		Status:     "bad request",
		StatusCode: http.StatusBadRequest,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			"X-Test-Header":  []string{"response"},
			"Content-Length": []string{"15"},
		},
		Body:          io.NopCloser(bytes.NewReader([]byte(`{"hello": "echo"}`))),
		ContentLength: 15,
	}

	clone, err := CloneResponseWithBody(res)
	require.NoError(t, err)

	RequireReaderContentsMatch(t, res.Body, clone.Body)
	require.EqualValues(t, res.Status, clone.Status)
	require.EqualValues(t, res.StatusCode, clone.StatusCode)
	require.EqualValues(t, res.Proto, clone.Proto)
	require.EqualValues(t, res.ProtoMajor, clone.ProtoMajor)
	require.EqualValues(t, res.ProtoMinor, clone.ProtoMinor)
	require.EqualValues(t, res.Header, clone.Header)
	require.EqualValues(t, res.ContentLength, clone.ContentLength)
	require.EqualValues(t, res.TransferEncoding, clone.TransferEncoding)
	require.EqualValues(t, res.Request, clone.Request)
}
