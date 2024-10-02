package spec

import (
	"bytes"
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/internaltools"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSerializeDeserializeRequest(t *testing.T) {
	expectedBody := bytes.NewBufferString("hello world")
	req, err := http.NewRequest("GET", "https://example.com",
		bytes.NewBuffer(expectedBody.Bytes()),
	)
	req.RequestURI = req.URL.String()
	req.TransferEncoding = []string{"chunked"}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	require.NoError(t, err)

	rec := NewRequestRecorder(req)
	_, _ = io.ReadAll(req.Body)
	sReq, err := SerializeRequest(rec)
	require.NoError(t, err)

	expected := &SerializableRequest{
		HTTPMessage: HTTPMessage{
			Header:           req.Header,
			Body:             "hello world",
			ContentType:      "text/plain; charset=utf-8",
			ContentLength:    11,
			TransferEncoding: req.TransferEncoding,
			Raw:              "GET https://example.com HTTP/1.1\r\nTransfer-Encoding: chunked\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nb\r\nhello world\r\n0\r\n\r\n",
		},
		Method: "GET",
		URL:    req.URL,
	}

	require.EqualValues(t, expected, sReq)

	dReq, err := DeserializeRequest(sReq)
	require.NoError(t, err)
	internaltools.RequireReaderContentsMatch(t, expectedBody, dReq.Body)
	req.Body = nil
	dReq.Body = nil
	req.GetBody = nil
	dReq.GetBody = nil

	require.EqualValues(t, req, dReq)
}

func TestSerializeAndDeserializeResponse(t *testing.T) {
	expectedBody := bytes.NewBufferString("hello world")
	res := &http.Response{
		StatusCode:       http.StatusOK,
		Proto:            "HTTP/1.1",
		Status:           http.StatusText(http.StatusOK),
		ProtoMajor:       1,
		ProtoMinor:       1,
		ContentLength:    11,
		TransferEncoding: []string{"chunked"},
		Header: http.Header{
			"Content-Type":      []string{"text/plain; charset=utf-8"},
			"Transfer-Encoding": []string{"chunked"},
		},
		Body: io.NopCloser(bytes.NewBuffer(expectedBody.Bytes())),
	}

	writer := NewResponseRecorder(httptest.NewRecorder())
	writer.WriteHeader(res.StatusCode)
	writer.Header().Set("Content-Type", res.Header.Get("Content-Type"))
	writer.Header().Set("Transfer-Encoding", res.Header.Get("Transfer-Encoding"))

	_, err := writer.Write(expectedBody.Bytes())
	require.NoError(t, err)

	sRes, err := SerializeResponse(writer)
	require.NoError(t, err)

	expected := &SerializableResponse{
		HTTPMessage: HTTPMessage{
			Header:           res.Header,
			Body:             "hello world",
			ContentType:      "text/plain; charset=utf-8",
			ContentLength:    11,
			TransferEncoding: res.TransferEncoding,
			Raw:              "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nb\r\nhello world\r\n0\r\n\r\n",
		},
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
	}

	require.EqualValues(t, expected, sRes)

	dRes, err := DeserializeResponse(sRes, nil)
	require.NoError(t, err)
	internaltools.RequireReaderContentsMatch(t, res.Body, dRes.Body)
	res.Body = nil
	dRes.Body = nil

	require.EqualValues(t, res, dRes)
}
