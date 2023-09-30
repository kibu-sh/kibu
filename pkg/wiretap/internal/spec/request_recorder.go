package spec

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

var _ RequestCloner = (*RequestRecorder)(nil)

type RequestRecorder struct {
	req        *http.Request
	bodyBuffer *bytes.Buffer
	teeReader  io.Reader
}

func (r RequestRecorder) Body() []byte {
	return r.bodyBuffer.Bytes()
}

func (r RequestRecorder) Clone() *http.Request {
	u := *r.req.URL
	req := r.req.Clone(context.Background())
	req.Body = io.NopCloser(bytes.NewBuffer(r.bodyBuffer.Bytes()))

	if u.Host == "" {
		u.Host = r.req.Host
	}

	if u.Scheme == "" {
		if r.req.TLS != nil {
			u.Scheme = "https"
		} else {
			u.Scheme = "http"
		}
	}

	req.URL = &u
	req.RequestURI = u.String()
	return req
}

func NewRequestRecorder(r *http.Request) *RequestRecorder {
	buf := bytes.NewBuffer(nil)
	r.Body = io.NopCloser(io.TeeReader(r.Body, buf))
	return &RequestRecorder{
		req:        r.Clone(context.Background()),
		bodyBuffer: buf,
	}
}

var _ RequestCloner = (*MultiReadRequestCloner)(nil)

type MultiReadRequestCloner struct {
	Req *http.Request
}

func (r MultiReadRequestCloner) Body() []byte {
	body := bytes.NewBuffer(nil)
	if r.Req.Body == nil {
		r.Req.Body = http.NoBody
	}
	_, _ = body.ReadFrom(r.Req.Body)
	r.Req.Body = io.NopCloser(bytes.NewBuffer(body.Bytes()))
	return body.Bytes()
}

func (r MultiReadRequestCloner) Clone() *http.Request {
	req := r.Req.Clone(context.Background())
	req.Body = io.NopCloser(bytes.NewBuffer(r.Body()))
	return req
}
