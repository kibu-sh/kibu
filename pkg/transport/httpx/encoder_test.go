package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/discernhq/devx/pkg/transport"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

var _ transport.Response = (*mockTransportResponse)(nil)

type mockTransportResponse struct {
	mock.Mock
	headers http.Header
	buf     *bytes.Buffer
}

func (m *mockTransportResponse) GetStatusCode() int {
	return http.StatusOK
}

func (m *mockTransportResponse) BodyBuffer() *bytes.Buffer {
	return m.buf
}

func (m *mockTransportResponse) BytesWritten() int64 {
	return int64(m.buf.Len())
}

func (m *mockTransportResponse) Write(p []byte) (n int, err error) {
	return m.buf.Write(p)
}

func (m *mockTransportResponse) Headers() http.Header {
	m.Called()
	return m.headers
}

func (m *mockTransportResponse) SetStatusCode(i int) {
	m.Called(i)
}

func (m *mockTransportResponse) DelCookie(cookie http.Cookie) transport.Response {
	panic("should never be called")
}

func (m *mockTransportResponse) DelCookieByName(name string) transport.Response {
	panic("should never be called")
}

func (m *mockTransportResponse) SetCookie(cookie http.Cookie) transport.Response {
	panic("should never be called")
}

func (m *mockTransportResponse) Redirect(req transport.Request, url string, code int) {
	panic("should never be called")
}

func (m *mockTransportResponse) Underlying() any {
	panic("should never be called")
}

func TestJSONErrorEncoder(t *testing.T) {
	ctx := context.Background()
	encoder := JSONErrorEncoder()
	t.Run("should properly encode wrapped transport error", func(t *testing.T) {
		resp := &mockTransportResponse{
			headers: http.Header{},
			buf:     new(bytes.Buffer),
		}
		expected := DefaultJSONError{
			Message: "Testing",
			Status:  http.StatusNotFound,
		}
		expectedBytes, err := json.Marshal(expected)
		require.NoError(t, err)
		resp.On("SetStatusCode", http.StatusNotFound).Return()
		resp.On("Headers").Return(http.Header{})
		err = encoder(ctx, resp, errors.Join(errors.New("testing"), expected))
		require.NoError(t, err)
		require.JSONEq(t, string(expectedBytes), resp.buf.String())
	})

	t.Run("should properly use default error message", func(t *testing.T) {
		resp := &mockTransportResponse{
			headers: http.Header{},
			buf:     new(bytes.Buffer),
		}
		resp.On("SetStatusCode", http.StatusInternalServerError).Return()
		resp.On("Headers").Return(http.Header{})
		err := encoder(ctx, resp, errors.New("broken"))
		require.NoError(t, err)
		require.Contains(t, resp.buf.String(), "broken")
		require.Equal(t, resp.Headers().Get("Content-Type"), "application/json")
	})
}
