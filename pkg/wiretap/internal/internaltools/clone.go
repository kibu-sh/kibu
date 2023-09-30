package internaltools

import (
	"bytes"
	"compress/gzip"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

var (
	ErrNilRequest  = errors.New("cannot clone nil request")
	ErrNilResponse = errors.New("cannot clone nil response")
)

func CloneRequestWithBody(originalRequest *http.Request) (clonedRequest *http.Request, err error) {
	if originalRequest == nil {
		return nil, ErrNilRequest
	}
	clonedRequest = originalRequest.Clone(originalRequest.Context())
	clonedRequest.Body, err = CloneRequestBody(originalRequest)
	return
}

func CloneResponseWithBody(originalResponse *http.Response) (clonedResponse *http.Response, err error) {
	if originalResponse == nil {
		return nil, ErrNilResponse
	}
	clonedResponse = new(http.Response)
	*clonedResponse = *originalResponse
	clonedResponse.Body, err = CloneResponseBody(originalResponse)
	return
}

func NoResponseBody(originalResponse *http.Response) bool {
	return BodyIsNil(originalResponse.Body)
}

func HasRequestBody(originalRequest *http.Request) bool {
	return !NoRequestBody(originalRequest)
}

func NoRequestBody(originalRequest *http.Request) bool {
	return BodyIsNil(originalRequest.Body)
}

func HasResponseBody(originalResponse *http.Response) bool {
	return !NoResponseBody(originalResponse)
}

func BodyIsNil(body io.ReadCloser) bool {
	return body == http.NoBody || body == nil
}

func CloneRequestBody(originalRequest *http.Request) (bodyCopy io.ReadCloser, err error) {
	bodyCopy = http.NoBody

	var bodyBuffer bytes.Buffer
	if NoRequestBody(originalRequest) {
		return
	}

	_, err = bodyBuffer.ReadFrom(originalRequest.Body)
	if err != nil {
		return
	}

	bodyCopy = io.NopCloser(bytes.NewReader(bodyBuffer.Bytes()))
	originalRequest.Body = io.NopCloser(bytes.NewReader(bodyBuffer.Bytes()))
	return
}

func CloneRequestBodyAsBuffer(originalRequest *http.Request) (bodyCopy *bytes.Buffer, err error) {
	bodyCopy = new(bytes.Buffer)
	reader, err := CloneRequestBody(originalRequest)
	if err != nil {
		return
	}
	_, err = bodyCopy.ReadFrom(reader)
	return
}

func CloneResponseBody(originalResponse *http.Response) (bodyCopy io.ReadCloser, err error) {
	bodyCopy = http.NoBody

	var bodyBuffer bytes.Buffer
	if NoResponseBody(originalResponse) {
		return
	}

	_, err = bodyBuffer.ReadFrom(originalResponse.Body)
	if err != nil {
		return
	}

	bodyCopy = io.NopCloser(bytes.NewReader(bodyBuffer.Bytes()))
	originalResponse.Body = io.NopCloser(bytes.NewReader(bodyBuffer.Bytes()))
	return
}

func RequireReaderContentsMatch(t *testing.T, expected io.Reader, actual io.Reader) {
	expectedBuffer := new(bytes.Buffer)
	_, err := expectedBuffer.ReadFrom(expected)
	require.NoError(t, err)

	actualBuffer := new(bytes.Buffer)
	_, err = actualBuffer.ReadFrom(actual)
	require.NoError(t, err)

	require.Equal(t, string(expectedBuffer.Bytes()), string(actualBuffer.Bytes()))

	return
}

func ReadBodyIfPresent(bodyReader io.ReadCloser) ([]byte, error) {
	if !BodyIsNil(bodyReader) {
		return io.ReadAll(bodyReader)
	}
	return nil, nil
}

func GzipUnwrapReader(bodyReader io.ReadCloser, header http.Header) (io.ReadCloser, error) {
	var err error
	if header.Get("Content-Encoding") == "gzip" {
		header.Del("Content-Encoding")
		bodyReader, err = gzip.NewReader(bodyReader)
		if err != nil {
			return nil, err
		}
	}
	return bodyReader, err
}

// DrainBody reads all of b to memory and then returns two equivalent
// ReadClosers yielding the same bytes.
//
// It returns an error if the initial slurp of all bytes fails. It does not attempt
// to make the returned ReadClosers have identical error-matching behavior.
func DrainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == nil || b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return io.NopCloser(&buf), io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
