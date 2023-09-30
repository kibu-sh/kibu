package proxy

import (
	"errors"
	"fmt"
	"github.com/discernhq/devx/pkg/wiretap/internal/internaltools"
	"github.com/discernhq/devx/pkg/wiretap/internal/spec"
	"net/http"
)

var _ http.RoundTripper = &ReplayTransport{}
var (
	ErrNoRouteMatch = errors.New("no match found")
	ErrRouteMatch   = errors.New("router error")
	ErrStoreRead    = errors.New("store read error")
)

type ReplayTransport struct {
	Router spec.SnapshotRouter
	Store  spec.SnapshotStore
}

func NewReplayTransport(router spec.SnapshotRouter, store spec.SnapshotStore) *ReplayTransport {
	return &ReplayTransport{
		Router: router,
		Store:  store,
	}
}

func NewReplayClient(transport *ReplayTransport) *http.Client {
	return &http.Client{
		Transport: transport,
	}
}

func (t *ReplayTransport) RoundTrip(request *http.Request) (res *http.Response, err error) {
	if request.Body != nil {
		// TODO: TEST
		// simulate a body read to ensure the upstream middleware captures the body
		// this will read and clone the body
		_, _ = internaltools.CloneRequestBody(request)
	}

	ref, err := t.Router.Match(request)
	if err != nil {
		// TODO: TEST THIS
		err = errors.Join(ErrRouteMatch, err, errFromRequest(request))
		return
	}

	if ref == nil {
		err = errors.Join(ErrNoRouteMatch, err, errFromRequest(request))
		return
	}

	snapshot, err := t.Store.Read(ref)
	if err != nil {
		err = errors.Join(ErrStoreRead, err, errFromRequest(request))
		return
	}

	if snapshot.Response == nil {
		err = errors.Join(ErrNoRouteMatch, err, errFromRequest(request))
		return
	}

	templateVars, err := spec.RequestToTemplate(request)
	if err != nil {
		return
	}

	res, err = spec.DeserializeResponse(snapshot.Response, map[string]any{
		"Request": templateVars,
	})
	if err != nil {
		return
	}

	res.Header.Add("X-Wiretap-Cache-Hit", "true")

	return
}

func errFromRequest(request *http.Request) error {
	return fmt.Errorf("%s %s", request.Method, request.URL.String())
}
