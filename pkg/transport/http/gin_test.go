package http

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestGin(t *testing.T) {
	svc := testSvc{}
	h := NewHandler("/home/:name", svc.Call)
	s := NewServer(NewGinRouter(), []Controller{h})
	require.HTTPStatusCode(t, http.HandlerFunc(s.ServeHTTP), "GET", "/home/test", nil, http.StatusOK)
	require.HTTPBodyContains(t, http.HandlerFunc(s.ServeHTTP), "GET", "/home/test", nil, "test")
	require.HTTPStatusCode(t, http.HandlerFunc(s.ServeHTTP), "GET", "/example", nil, http.StatusNotFound)
}
