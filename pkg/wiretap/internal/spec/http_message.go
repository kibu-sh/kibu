package spec

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type HTTPMessage struct {
	Header           http.Header
	Body             string
	ContentType      string
	ContentLength    int64
	TransferEncoding []string
	Raw              string
}

type SerializableRequest struct {
	HTTPMessage
	Method string
	URL    *url.URL
}

type SerializableResponse struct {
	HTTPMessage
	StatusCode int
	Status     string
}

func DeserializeRequest(s *SerializableRequest) (*http.Request, error) {
	req, err := http.NewRequest(s.Method, s.URL.String(), bodyFromString(s.Body))
	if err != nil {
		return nil, err
	}

	req.Header = s.Header
	req.TransferEncoding = s.TransferEncoding
	req.ContentLength = s.ContentLength
	req.RequestURI = req.URL.String()

	return req, nil
}

func bodyFromString(rawBody string) io.ReadCloser {
	if rawBody == "" {
		return http.NoBody
	} else {
		return io.NopCloser(strings.NewReader(rawBody))
	}
}

func DeserializeResponse(s *SerializableResponse, templateVars map[string]any) (*http.Response, error) {
	res := &http.Response{
		Status:           s.Status,
		StatusCode:       s.StatusCode,
		Proto:            "HTTP/1.1",
		ProtoMajor:       1,
		ProtoMinor:       1,
		ContentLength:    s.ContentLength,
		TransferEncoding: s.TransferEncoding,
		Header:           s.Header,
		Body:             bodyFromString(s.Body),
	}

	if templateVars == nil {
		return res, nil
	}

	tmp, err := BodyTemplate().Parse(s.Body)
	if err != nil {
		return nil, err
	}

	buff := bytes.NewBuffer(nil)
	err = tmp.Execute(buff, templateVars)
	if err != nil {
		return nil, err
	}

	res.Body = bodyFromString(buff.String())

	return res, nil
}

func SerializeRequest(r RequestCloner) (*SerializableRequest, error) {
	req := r.Clone()
	body := r.Body()
	rawReq, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}

	return &SerializableRequest{
		HTTPMessage: HTTPMessage{
			Header:           req.Header,
			Body:             string(body),
			Raw:              string(rawReq),
			ContentLength:    int64(len(body)),
			TransferEncoding: req.TransferEncoding,
			ContentType:      req.Header.Get("Content-Type"),
		},
		Method: req.Method,
		URL:    req.URL,
	}, nil
}

type ResponseCloner interface {
	Body() []byte
	Clone() *http.Response
}

type RequestCloner interface {
	Body() []byte
	Clone() *http.Request
}

func SerializeResponse(rc ResponseCloner) (*SerializableResponse, error) {
	res := rc.Clone()
	body := rc.Body()
	rawRes, err := httputil.DumpResponse(res, true)
	if err != nil {
		return nil, err
	}

	return &SerializableResponse{
		Status:     res.Status,
		StatusCode: res.StatusCode,
		HTTPMessage: HTTPMessage{
			Header:           res.Header,
			Body:             string(body),
			Raw:              string(rawRes),
			ContentLength:    int64(len(body)),
			ContentType:      res.Header.Get("Content-Type"),
			TransferEncoding: res.TransferEncoding,
		},
	}, nil
}
