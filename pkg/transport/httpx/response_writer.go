package httpx

import (
	"bytes"
	"github.com/discernhq/devx/pkg/transport"
	"io"
	"net/http"
	"time"
)

var _ transport.Response = (*ResponseWriter)(nil)
var _ io.Writer = (*ResponseWriter)(nil)

type ResponseWriter struct {
	http.ResponseWriter
	bytesWritten   int64
	sentStatusCode int
	bodyBuffer     *bytes.Buffer
	multiWriter    io.Writer
}

func (r *ResponseWriter) BodyBuffer() *bytes.Buffer {
	return r.bodyBuffer
}

func (r *ResponseWriter) Write(b []byte) (int, error) {
	written, err := r.multiWriter.Write(b)
	r.bytesWritten += int64(written)

	// if the status code has not been set, default to 200
	// this is implied on the first write of the response
	if r.sentStatusCode == 0 {
		r.sentStatusCode = http.StatusOK
	}

	return written, err
}

func (r *ResponseWriter) BytesWritten() int64 {
	return r.bytesWritten
}

func (r *ResponseWriter) DelCookie(cookie http.Cookie) transport.Response {
	cookie.Value = ""
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Expires = time.Unix(0, 0)
	return r.SetCookie(cookie)
}

func (r *ResponseWriter) DelCookieByName(name string) transport.Response {
	return r.DelCookie(http.Cookie{
		Name: name,
	})
}

func (r *ResponseWriter) SetCookie(cookie http.Cookie) transport.Response {
	http.SetCookie(r.ResponseWriter, &cookie)
	return r
}

func (r *ResponseWriter) Redirect(req transport.Request, url string, code int) {
	http.Redirect(r.ResponseWriter, req.Underlying().(*http.Request), url, code)
}

func (r *ResponseWriter) SetStatusCode(i int) {
	r.ResponseWriter.WriteHeader(i)
	r.sentStatusCode = i
}

func (r *ResponseWriter) GetStatusCode() int {
	return r.sentStatusCode
}

func (r *ResponseWriter) Headers() http.Header {
	return r.ResponseWriter.Header()
}

func (r *ResponseWriter) Underlying() any {
	return r.ResponseWriter
}

func NewResponse(w http.ResponseWriter) *ResponseWriter {
	resWriter := &ResponseWriter{
		ResponseWriter: w,
		bodyBuffer:     new(bytes.Buffer),
	}
	resWriter.multiWriter = io.MultiWriter(w, resWriter.bodyBuffer)
	return resWriter
}
