package httpx

import (
	"github.com/discernhq/devx/pkg/transport"
	"net/http"
	"time"
)

var _ transport.Response = (*Response)(nil)

type Response struct {
	http.ResponseWriter
}

func (r *Response) DelCookie(cookie http.Cookie) transport.Response {
	cookie.Value = ""
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Expires = time.Unix(0, 0)
	return r.SetCookie(cookie)
}

func (r *Response) DelCookieByName(name string) transport.Response {
	return r.DelCookie(http.Cookie{
		Name: name,
	})
}

func (r *Response) SetCookie(cookie http.Cookie) transport.Response {
	http.SetCookie(r.ResponseWriter, &cookie)
	return r
}

func (r *Response) Redirect(req transport.Request, url string, code int) {
	http.Redirect(r.ResponseWriter, req.Underlying().(*http.Request), url, code)
}

func (r *Response) SetStatusCode(i int) {
	r.ResponseWriter.WriteHeader(i)
}

func (r *Response) Headers() http.Header {
	return r.ResponseWriter.Header()
}

func (r *Response) Underlying() any {
	return r.ResponseWriter
}

func NewResponse(w http.ResponseWriter) *Response {
	return &Response{w}
}
