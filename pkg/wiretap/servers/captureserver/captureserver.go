package captureserver

import (
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/spec"
	"github.com/kibu-sh/kibu/pkg/wiretap/proxy"
	"net/http"
)

func BindToMux(
	mux *http.ServeMux,
	topic spec.SnapshotMessageTopic,
) *http.ServeMux {
	mux.Handle("/", proxy.NewCaptureProxy(topic))
	return mux
}
