package spec

import (
	"bytes"
	"io"
	"maps"
	"net/http"
)

var _ ResponseCloner = (*ResponseRecorder)(nil)
var _ http.ResponseWriter = (*ResponseRecorder)(nil)

type ResponseRecorder struct {
	statusCode   int
	bytesWritten int64
	req          *http.Request
	res          http.ResponseWriter
	bodyBuffer   *bytes.Buffer
	multiWriter  io.Writer
	headerMap    http.Header
	snapHeader   http.Header
}

func (r *ResponseRecorder) Body() []byte {
	return r.bodyBuffer.Bytes()
}

func (r *ResponseRecorder) Clone() *http.Response {
	header := maps.Clone(r.snapHeader)
	return &http.Response{
		Status:           http.StatusText(r.statusCode),
		StatusCode:       r.statusCode,
		Proto:            "HTTP/1.1",
		ProtoMajor:       1,
		ProtoMinor:       1,
		ContentLength:    r.bytesWritten,
		Header:           header,
		TransferEncoding: header.Values("Transfer-Encoding"),
		Body:             io.NopCloser(bytes.NewBuffer(r.bodyBuffer.Bytes())),
	}
}

func (r *ResponseRecorder) Header() http.Header {
	return r.headerMap
}

func (r *ResponseRecorder) Write(b []byte) (int, error) {
	written, err := r.multiWriter.Write(b)
	r.bytesWritten += int64(written)

	// if the status code has not been set, default to 200
	// this is implied on the first write of the response
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}

	if r.headerMap == nil {
		r.headerMap = make(http.Header)
	}
	r.snapHeader = r.headerMap.Clone()
	return written, err
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.res.WriteHeader(statusCode)
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	bodyBuffer := bytes.NewBuffer(nil)
	return &ResponseRecorder{
		res:         w,
		bodyBuffer:  bodyBuffer,
		multiWriter: io.MultiWriter(w, bodyBuffer),
		headerMap:   w.Header(),
	}
}

var _ ResponseCloner = (*MultiReadResponseCloner)(nil)

type MultiReadResponseCloner struct {
	Res *http.Response
}

func (r MultiReadResponseCloner) Body() []byte {
	body := bytes.NewBuffer(nil)
	if r.Res.Body == nil {
		r.Res.Body = http.NoBody
	}
	_, _ = body.ReadFrom(r.Res.Body)
	r.Res.Body = io.NopCloser(bytes.NewBuffer(body.Bytes()))
	return body.Bytes()
}

func (r MultiReadResponseCloner) Clone() *http.Response {
	c := *r.Res
	c.Body = io.NopCloser(bytes.NewBuffer(r.Body()))
	return &c
}
