package httpx

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/transport"
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
	e := transport.NewEndpoint(svc.Call)
	h := NewHandler("/", e)
	require.HTTPStatusCode(t, http.HandlerFunc(h.ServeHTTP), "GET", "/example", nil, http.StatusOK)
}
