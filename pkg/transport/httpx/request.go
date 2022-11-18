package httpx

import (
	"github.com/discernhq/devx/pkg/transport"
	"io"
	"mime"
	"net/http"
	"net/url"
)

var _ transport.Request = (*Request)(nil)

type Request struct {
	*http.Request
}

func (r *Request) URL() *url.URL {
	return r.Request.URL
}

// TODO: path can be incorrect when behind a proxy
// TODO: what about the original mounting path? maybe we don't care (wait until someone asks)

func (r *Request) Path() string {
	return r.URL().Path
}

func (r *Request) Body() io.ReadCloser {
	return r.Request.Body
}

func (r *Request) ParseMediaType() (mediatype string, params map[string]string, err error) {
	return mime.ParseMediaType(r.Header.Get("Content-Type"))
}

func (r *Request) Underlying() any {
	return r.Request
}

func (r *Request) Method() string {
	return r.Request.Method
}

func (r *Request) PathParams() url.Values {
	return PathParamsFromContext(r.Request.Context())
}

func (r *Request) QueryParams() url.Values {
	return r.Request.URL.Query()
}

func (r *Request) Headers() http.Header {
	return r.Request.Header
}

func NewRequest(r *http.Request) *Request {
	return &Request{r}
}
