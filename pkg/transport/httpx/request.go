package httpx

import (
	"bytes"
	"context"
	"github.com/discernhq/devx/pkg/transport"
	"io"
	"mime"
	"net/http"
	"net/url"
)

var _ transport.Request = (*Request)(nil)

type Request struct {
	*http.Request

	bodyBuffer *bytes.Buffer
}

func (r *Request) BodyBuffer() *bytes.Buffer {
	return r.bodyBuffer
}

func (r *Request) Version() string {
	return r.Request.Proto
}

func (r *Request) WithContext(ctx context.Context) transport.Request {
	r.Request = r.Request.WithContext(ctx)
	return r
}

func (r *Request) URL() *url.URL {
	if !r.Request.URL.IsAbs() {
		r.Request.URL.Scheme = "http"
		r.Request.URL.Host = r.Request.Host
		if r.Request.TLS != nil || r.Request.Header.Get("X-Forwarded-Proto") == "https" {
			r.Request.URL.Scheme = "https"
		}
	}
	return r.Request.URL
}

// TODO: path can be incorrect when behind a proxy
// TODO: what about the original mounting path? maybe we don't care (wait until someone asks)

func (r *Request) Path() string {
	return r.URL().Path
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
	req := &Request{
		Request:    r,
		bodyBuffer: new(bytes.Buffer),
	}
	r.Body = newTeeReadCloser(r.Body, req.bodyBuffer)
	return req
}

var _ io.ReadCloser = (*teeReadCloser)(nil)

type teeReadCloser struct {
	original io.ReadCloser
	tee      io.Reader
}

func (t teeReadCloser) Read(p []byte) (n int, err error) {
	return t.tee.Read(p)
}

func (t teeReadCloser) Close() error {
	return t.original.Close()
}

func newTeeReadCloser(original io.ReadCloser, writer io.Writer) *teeReadCloser {
	return &teeReadCloser{
		original: original,
		tee:      io.TeeReader(original, writer),
	}
}

func (r *Request) Body() io.ReadCloser {
	return r.Request.Body
}
