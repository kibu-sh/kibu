package httpx

import (
	"github.com/discernhq/devx/pkg/transport"
	"net/http"
)

var _ transport.Response = (*Response)(nil)

type Response struct {
	http.ResponseWriter
}

func (r Response) SetStatusCode(i int) {
	r.ResponseWriter.WriteHeader(i)
}

func (r Response) Headers() http.Header {
	return r.ResponseWriter.Header()
}

func (r Response) Underlying() any {
	return r.ResponseWriter
}

func NewResponse(w http.ResponseWriter) *Response {
	return &Response{w}
}
