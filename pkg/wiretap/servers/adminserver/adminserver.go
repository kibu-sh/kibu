package adminserver

import (
	"context"
	"github.com/discernhq/devx/pkg/wiretap/internal/spec"
	"github.com/discernhq/devx/pkg/wiretap/ui"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"log/slog"
	"net/http"
	"net/url"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"strings"
	"time"
)

var (
	AdminPrefix    = "/__admin/"
	APIPrefix      = mustJoinURLPath(AdminPrefix, "/api/")
	UIPrefix       = mustJoinURLPath(AdminPrefix, "/ui/")
	SnapshotPrefix = mustJoinURLPath(APIPrefix, "/snapshot/")

	Endpoints = struct {
		Base           string
		API            string
		UI             string
		Echo           string
		SnapshotStream string
	}{
		Base:           AdminPrefix,
		API:            APIPrefix,
		UI:             UIPrefix,
		Echo:           mustJoinURLPath(APIPrefix, "echo"),
		SnapshotStream: mustJoinURLPath(SnapshotPrefix, "stream"),
	}
)

func mustJoinURLPath(base string, paths ...string) string {
	return lo.Must(url.JoinPath(base, paths...))
}

type AdminServer struct {
	snapshotTopic spec.SnapshotMessageTopic
	logger        *slog.Logger
}

func NewAdminServer(snapshotTopic spec.SnapshotMessageTopic) *AdminServer {
	return &AdminServer{
		snapshotTopic: snapshotTopic,
		logger:        slog.Default(),
	}
}

func (s *AdminServer) SnapshotStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := s.logger.With("remote", r.RemoteAddr)
	log.InfoContext(ctx, "snapshot stream connection opened")

	ws, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.ErrorContext(ctx, "failed to accept websocket connection", slog.String("error", err.Error()))
		return
	}

	defer func() {
		var level slog.Level = slog.LevelInfo
		var message = "connection closed"
		var status websocket.StatusCode = http.StatusOK
		if err != nil && !errors.Is(err, context.Canceled) {
			level = slog.LevelError
			status = http.StatusInternalServerError
			message = strings.Join([]string{
				message,
				err.Error(),
			}, ":")
		}
		_ = ws.Close(status, message)
		log.LogAttrs(ctx, level, message)
	}()

	ctx = ws.CloseRead(ctx)
	stream, err := s.snapshotTopic.Subscribe(ctx)
	if err != nil {
		err = errors.Wrap(err, "failed to subscribe to snapshot stream")
		return
	}

	defer func() {
		stream.Unsubscribe()
		log.DebugContext(ctx, "unsubscribe from snapshot stream")
	}()

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case snapshot, more := <-stream.Channel():
			if !more {
				log.InfoContext(ctx, "snapshot stream closed")
				return
			}

			if writeErr := writeTimeout(ctx, ws, snapshot); writeErr != nil {
				log.LogAttrs(ctx, slog.LevelError, "failed to send snapshot",
					slog.String("error", writeErr.Error()),
					slog.String("snapshot", snapshot.ID),
				)
				continue
			}

			log.InfoContext(ctx, "sent snapshot", slog.String("snapshot", snapshot.ID))
		}
	}
}

func writeTimeout(ctx context.Context, ws *websocket.Conn, snapshot spec.Snapshot) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	return wsjson.Write(ctx, ws, snapshot)
}

func BindToMux(mux *http.ServeMux, topic spec.SnapshotMessageTopic) *http.ServeMux {
	adminServer := NewAdminServer(topic)
	mux.Handle(UIPrefix, http.StripPrefix(UIPrefix, ui.Handler()))
	mux.HandleFunc(Endpoints.SnapshotStream, adminServer.SnapshotStream)
	return mux
}
