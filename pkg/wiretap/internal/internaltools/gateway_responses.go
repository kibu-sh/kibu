package internaltools

import (
	"net/http"
)

func NewBadGatewayResponse(req *http.Request) *http.Response {
	return &http.Response{
		Status:     http.StatusText(http.StatusBadGateway),
		StatusCode: http.StatusBadGateway,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Body:       http.NoBody,
		Request:    req,
	}
}

func NewGatewayTimeoutResponse(req *http.Request) *http.Response {
	return &http.Response{
		Status:     http.StatusText(http.StatusGatewayTimeout),
		StatusCode: http.StatusGatewayTimeout,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Body:       http.NoBody,
		Request:    req,
	}
}
