package httpx

import (
	"github.com/discernhq/devx/pkg/transport"
	"net/http"
	"time"
)

var _ transport.Response = (*ResponseWriter)(nil)

type ResponseWriter struct {
	http.ResponseWriter
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
}

func (r *ResponseWriter) Headers() http.Header {
	return r.ResponseWriter.Header()
}

func (r *ResponseWriter) Underlying() any {
	return r.ResponseWriter
}

func NewResponse(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w}
}
