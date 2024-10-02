package httpx

import (
	"github.com/kibu-sh/kibu/pkg/transport"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestGin(t *testing.T) {
	svc := testSvc{}
	e := transport.NewEndpoint(svc.Call)
	h := NewHandler("/home/:name", e)
	m := NewGinMux()
	m.Handle(h)
	require.HTTPStatusCode(t, http.HandlerFunc(m.ServeHTTP), "GET", "/home/test", nil, http.StatusOK)
	require.HTTPBodyContains(t, http.HandlerFunc(m.ServeHTTP), "GET", "/home/test", nil, "test")
	require.HTTPStatusCode(t, http.HandlerFunc(m.ServeHTTP), "GET", "/example", nil, http.StatusNotFound)
}
