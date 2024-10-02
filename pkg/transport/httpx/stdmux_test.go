package httpx

import (
	"github.com/kibu-sh/kibu/pkg/transport"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	svc := testSvc{}
	h := NewHandler("/home", transport.NewEndpoint(svc.Call))
	m := NewStdLibMux()
	m.Handle(h)
	require.HTTPStatusCode(t, http.HandlerFunc(m.ServeHTTP), "GET", "/home", nil, http.StatusOK)
	require.HTTPStatusCode(t, http.HandlerFunc(m.ServeHTTP), "GET", "/example", nil, http.StatusNotFound)
}
