package replayserver

import (
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/spec"
	"github.com/kibu-sh/kibu/pkg/wiretap/proxy"
	"net/http"
)

func BindToMux(mux *http.ServeMux, store spec.SnapshotStore, router spec.SnapshotRouter, topic spec.SnapshotMessageTopic) *http.ServeMux {
	mux.Handle("/", proxy.NewReplayProxy(
		topic, proxy.NewReplayTransport(router, store),
	))
	return mux
}
