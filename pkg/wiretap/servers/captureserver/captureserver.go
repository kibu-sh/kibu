package captureserver

import (
	"github.com/discernhq/devx/pkg/wiretap/internal/spec"
	"github.com/discernhq/devx/pkg/wiretap/proxy"
	"net/http"
)

func BindToMux(
	mux *http.ServeMux,
	topic spec.SnapshotMessageTopic,
) *http.ServeMux {
	mux.Handle("/", proxy.NewCaptureProxy(topic))
	return mux
}
