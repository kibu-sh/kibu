package httpx

import (
	"github.com/discernhq/devx/pkg/transport"
	"net/http"
)

var _ transport.ResponseWriter = (*ResponseWriter)(nil)

type ResponseWriter struct {
	http.ResponseWriter
}

func (r ResponseWriter) Underlying() any {
	return r.ResponseWriter
}
