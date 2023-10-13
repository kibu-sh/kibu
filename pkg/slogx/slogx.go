package slogx

import (
	"bytes"
	"fmt"
	"github.com/discernhq/devx/pkg/transport"
	"github.com/pkg/errors"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type BuilderFunc func(logger *slog.Logger) *slog.Logger

type httpResponseLogRecord struct {
	StatusCode   int         `json:"status_code"`
	Headers      http.Header `json:"headers"`
	BytesWritten int64       `json:"bytes"`
	Body         string      `json:"body,omitempty"`
}

type httpRequestLogRecord struct {
	Version     string      `json:"proto"`
	Method      string      `json:"method"`
	URL         string      `json:"url"`
	Headers     http.Header `json:"headers"`
	Referer     string      `json:"referer"`
	UserAgent   string      `json:"useragent"`
	ClientIP    string      `json:"client_ip"`
	Host        string      `json:"host"`
	ContentType string      `json:"content_type"`
	Body        string      `json:"body,omitempty"`
}

func BindLogBuilders(logger *slog.Logger, builders ...BuilderFunc) *slog.Logger {
	for _, builder := range builders {
		logger = builder(logger)
	}
	return logger
}

func WithDuration(duration time.Duration) BuilderFunc {
	return func(logger *slog.Logger) *slog.Logger {
		return logger.With("duration", duration.Nanoseconds())
	}
}

func WithErrorInfo(err error) BuilderFunc {
	return func(logger *slog.Logger) *slog.Logger {
		if err == nil {
			return logger
		}

		// TODO: stack trace capture
		// TODO: unwrap errors recursively to log
		//https://pkg.go.dev/github.com/pkg/errors#hdr-Adding_context_to_an_error

		log := logger.
			With("error.message", err).
			With("error.kind", fmt.Sprintf("%T", err))

		if stackTrace, ok := buildErrorStackTrace(err); ok {
			log = log.With("error.stack", stackTrace)
		}

		return log
	}
}

func WithRequestInfo(r transport.Request) BuilderFunc {
	return func(logger *slog.Logger) *slog.Logger {
		return logger.With("http.request", httpRequestLogRecord{
			Version:     r.Version(),
			Method:      r.Method(),
			URL:         r.URL().String(),
			ContentType: r.Headers().Get("Content-Type"),
			Referer:     r.Headers().Get("Referer"),
			UserAgent:   r.Headers().Get("User-Agent"),
			Host:        r.Headers().Get("Host"),
			Headers:     normalizeHeaderKeys(r.Headers()),
			ClientIP:    chooseClientIPHeaderFromDefaults(r),
			Body:        logUpToMaxSize(r.BodyBuffer(), defaultMaxPayloadDumpSize),
		})
	}
}

// defaultMaxPayloadDumpSize is the default size 256KB
const defaultMaxPayloadDumpSize = 256 * 1024

func WithResponseInfo(w transport.Response) BuilderFunc {
	return func(logger *slog.Logger) *slog.Logger {
		return logger.
			With("http.response", httpResponseLogRecord{
				StatusCode:   w.GetStatusCode(),
				BytesWritten: w.BytesWritten(),
				Headers:      w.Headers(),
				Body:         logUpToMaxSize(w.BodyBuffer(), defaultMaxPayloadDumpSize),
			})
	}
}

func logUpToMaxSize(w *bytes.Buffer, max int) string {
	if w.Len() <= max {
		return w.String()
	}
	return fmt.Sprintf("payload omitted because body size %d exceeds max of %d", w.Len(), max)
}

func chooseClientIPHeader(r transport.Request, keys []string) (ip string) {
	for _, key := range keys {
		if ip = r.Headers().Get(key); ip != "" {
			return
		}
	}
	return
}

func chooseClientIPHeaderFromDefaults(r transport.Request) (ip string) {
	return chooseClientIPHeader(r, []string{
		"True-Client-IP",
		"CF-Connecting-IP",
		"X-Client-IP",
		"X-Real-Ip",
		"X-Forwarded-For",
	})
}

func normalizeHeaderKeys(headers http.Header) http.Header {
	var normalized http.Header = make(map[string][]string)
	for k, v := range headers {
		normalized[http.CanonicalHeaderKey(k)] = v
	}
	return normalized
}

func buildErrorStackTrace(err error) (string, bool) {
	tracer, ok := err.(stackTracer)
	if !ok {
		return "", false
	}

	stackTrace := new(strings.Builder)
	for _, f := range tracer.StackTrace() {
		stackTrace.WriteString(
			fmt.Sprintf("%+s:%d\n", f, f),
		)

	}

	return stackTrace.String(), true
}
