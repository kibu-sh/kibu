package httpx

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	svc := testSvc{}
	h := NewHandler("/home", svc.Call)
	s := NewServeMux(NewServeMuxParams{
		Router:      NewStdRouter(),
		Controllers: []Controller{h},
	})
	require.HTTPStatusCode(t, http.HandlerFunc(s.ServeHTTP), "GET", "/home", nil, http.StatusOK)
	require.HTTPStatusCode(t, http.HandlerFunc(s.ServeHTTP), "GET", "/example", nil, http.StatusNotFound)
}
