package internaltools

import (
	"context"
	"errors"
	"io"
	"net/http"
)

func RoundTripWithBadGateway(req *http.Request, transport http.RoundTripper) (res *http.Response, err error) {
	res, rtErr := transport.RoundTrip(req)
	switch {
	case errors.Is(rtErr, io.EOF):
		res = NewBadGatewayResponse(req)
	case errors.Is(rtErr, context.DeadlineExceeded):
		res = NewGatewayTimeoutResponse(req)
	default:
		err = rtErr
	}
	return
}
