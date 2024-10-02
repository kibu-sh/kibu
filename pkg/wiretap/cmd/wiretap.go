package main

import (
	"flag"
	"github.com/kibu-sh/kibu/pkg/wiretap"
	"github.com/kibu-sh/kibu/pkg/wiretap/certgen"
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/spec"
	"github.com/kibu-sh/kibu/pkg/wiretap/routers/dynamic"
	"github.com/kibu-sh/kibu/pkg/wiretap/rules/requestrules"
	"github.com/kibu-sh/kibu/pkg/wiretap/stores/archive"
	"github.com/pkg/errors"
	"log/slog"
	"os"
	"path/filepath"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
}

func main() {
	var err error
	log := slog.Default()
	defer func() {
		if err != nil {
			log.Error("exiting", "error", err.Error())
			os.Exit(1)
		}
	}()

	mode := flag.String("mode", "capture", "capture|replay")
	dir := flag.String("dir", "-", "directory where snapshots are written/read")

	flag.Parse()

	if *dir == "" {
		err = errors.New("must specify -dir")
		return
	}

	if *mode == "" {
		err = errors.New("must specify -mode")
		return
	}

	pool, err := certgen.LoadUserCachedCertPool()
	if err != nil {
		return
	}

	if err = mkTmpDirIfSpecified(dir); err != nil {
		return
	}

	if err = formatRelativePath(dir); err != nil {
		return
	}

	log.Debug("snapshot directory initialized at",
		"dir", *dir)

	snapshots, err := archive.LoadSnapshotsFromDir(*dir)
	if err != nil {
		return
	}

	var router spec.SnapshotRouter = dynamic.NewSnapshotRouter()

	for _, snapshot := range snapshots {
		router.Register(snapshot.Ref(), requestrules.BasicMatchRule(snapshot))
	}

	server := wiretap.NewServer().
		WithLogger(log).
		WithSnapshotDir(*dir).
		WithCertPool(pool).
		WithRouter(router)

	switch *mode {
	case "capture":
		server, err = server.StartInCaptureMode()
	case "replay":
		server, err = server.StartInReplayMode()
	default:
		err = errors.Errorf("invalid mode: %s", *mode)
	}

	if err != nil {
		return
	}

	log.Info("running in", "mode", *mode)
	err = server.Serve()
}

func formatRelativePath(dir *string) (err error) {
	if filepath.IsAbs(*dir) {
		return nil
	}

	*dir, err = filepath.Abs(*dir)
	if err != nil {
		return nil
	}

	return os.MkdirAll(*dir, 0755)
}

func mkTmpDirIfSpecified(dir *string) (err error) {
	if *dir != "-" {
		return
	}

	*dir, err = os.MkdirTemp("", spec.ToolName)
	if err != nil {
		return
	}
	return
}
