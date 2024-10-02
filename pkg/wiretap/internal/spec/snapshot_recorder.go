package spec

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type timeFunc func() time.Time

// NewSnapshotMiddleware returns a middleware that will capture snapshots of
// requests and responses and publish them to the provided topic.
// The middleware will also inject a timeFunc into the request context
// that will return the time the request was received.
// This is useful for testing and making the time deterministic.
// NOTE: This middleware will delete the Accept-Encoding header from the request
// This is because go's transport library will not decompress the response body
// if an upstream client requested compression.
func NewSnapshotMiddleware(topic SnapshotMessageTopic, getCurrentTime timeFunc) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			recorder, writer, req := NewSnapshotRecorder(writer, req, topic, getCurrentTime())
			defer recorder.Done()
			go recorder.Capture(getCurrentTime)

			// IMPORTANT
			// go's transport library won't decompress the response body
			// if an upstream client requested compression
			req.Header.Del("Accept-Encoding")
			next.ServeHTTP(writer, req)
		})
	}
}

// SnapshotRecorder is a recorder that will capture snapshots of requests and responses
// This is intended for use in http.Handler middleware.
// The caller must call Done() on the recorder when the request is complete.
// It is recommended to call Capture() in a goroutine to avoid blocking the request on publishing the snapshot.
// It is also recommended that you defer Done() to ensure it is called even if the request panics.
type SnapshotRecorder struct {
	res    *ResponseRecorder
	req    *RequestRecorder
	topic  SnapshotMessageTopic
	logger *slog.Logger
	start  time.Time
	end    time.Time
	wait   chan struct{}
}

func (r SnapshotRecorder) Capture(now func() time.Time) {
	<-r.wait
	r.end = now()
	sh, err := NewSnapshot(r.req, r.res, r.Duration())
	if err != nil {
		r.logger.Error("failed to create snapshot", slog.String("error", err.Error()))
		return
	}

	if err = r.topic.Publish(context.Background(), *sh); err != nil {
		r.logger.Error("failed to publish snapshot", slog.String("error", err.Error()))
	}
}

func (r SnapshotRecorder) Duration() time.Duration {
	return r.end.Sub(r.start)
}

func (r SnapshotRecorder) Done() {
	close(r.wait)
}

func NewSnapshotRecorder(
	w http.ResponseWriter,
	r *http.Request,
	topic SnapshotMessageTopic,
	start time.Time,
) (*SnapshotRecorder, http.ResponseWriter, *http.Request) {
	recorder := &SnapshotRecorder{
		res:    NewResponseRecorder(w),
		req:    NewRequestRecorder(r),
		logger: slog.Default(),
		topic:  topic,
		start:  start,
		wait:   make(chan struct{}),
	}

	return recorder, recorder.res, r
}
