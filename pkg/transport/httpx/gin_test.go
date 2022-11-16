package httpx

import (
	"github.com/discernhq/devx/pkg/transport"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestGin(t *testing.T) {
	svc := testSvc{}
	h := NewHandler("/home/:name", transport.NewController(
		transport.NewEndpoint(svc.Call),
	))
	m := NewGinMux()
	m.Handle(h)
	require.HTTPStatusCode(t, http.HandlerFunc(m.ServeHTTP), "GET", "/home/test", nil, http.StatusOK)
	require.HTTPBodyContains(t, http.HandlerFunc(m.ServeHTTP), "GET", "/home/test", nil, "test")
	require.HTTPStatusCode(t, http.HandlerFunc(m.ServeHTTP), "GET", "/example", nil, http.StatusNotFound)
}
