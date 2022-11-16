package httpx

import (
	"context"
	"github.com/discernhq/devx/pkg/transport"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

type testSvc struct {
}

type testReq struct {
	Name string `path:"name"`
}

type testRes struct {
	Name string `path:"name"`
}

func (s testSvc) Call(ctx context.Context, req testReq) (res testRes, err error) {
	return testRes(req), nil
}

func TestHandler_ServeHTTP(t *testing.T) {
	svc := testSvc{}
	h := NewHandler("/", transport.NewController(
		transport.NewEndpoint(svc.Call),
	))
	require.HTTPStatusCode(t, http.HandlerFunc(h.ServeHTTP), "GET", "/example", nil, http.StatusOK)
}
