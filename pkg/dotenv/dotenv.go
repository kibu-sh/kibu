package dotenv

import (
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"log/slog"
	"os"
	"path/filepath"
)

var ErrMaxDepthReached = errors.New("max depth reached")

func AutoLoadDotEnv(log *slog.Logger) {
	if log == nil {
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	log = log.With("module", "dotenv")

	cwd, _ := os.Getwd()
	single := os.Getenv("DOTENV_FILE")
	if single != "" {
		loadEnvFile(log, single)
		return
	}

	dir := os.Getenv("DOTENV_DIR")
	if dir != "" {
		cwd = dir
	}

	_ = searchAndLoadEnv(log, cwd, nil)
}

func searchAndLoadEnv(log *slog.Logger, currPath string, maxDepth *int) (err error) {
	currPath = filepath.Clean(currPath)
	if searchHasReachedMaxDepth(currPath, maxDepth) {
		err = errors.Wrapf(ErrMaxDepthReached, "at %s", currPath)
		log.Debug("max depth reached", "path", currPath)
		return
	}

	envPath := filepath.Join(currPath, ".env")
	if _, err := os.Stat(envPath); err == nil {
		loadEnvFile(log, envPath)
	} else {
		log.Debug("no .env file found in", "path", currPath)
	}

	if maxDepth != nil {
		*maxDepth--
	}

	// Recursive call: Move to the parent directory
	parentPath := filepath.Dir(currPath)
	return searchAndLoadEnv(log, parentPath, maxDepth)
}

func loadEnvFile(log *slog.Logger, path string) {
	log.Debug("loading .env", "path", path)
	file, err := os.Open(path)
	if err != nil {
		log.Error("failed to load .env",
			"path", path,
			"error", err.Error())
		return
	}

	envMap, err := godotenv.Parse(file)
	if err != nil {
		log.Error("failed to load .env",
			"path", path,
			"error", err.Error())
		return
	}

	for key, value := range envMap {
		_ = os.Setenv(key, value)
	}
}

func searchHasReachedMaxDepth(currPath string, maxDepth *int) bool {
	return (maxDepth != nil && *maxDepth <= 0) ||
		currPath == "/" ||
		currPath == "." ||
		currPath == ""
}
