package dotenv

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

const (
	parentShared = "KIBU_PARENT_SHARED"
	parentOnly   = "KIBU_PARENT_ONLY"
)

var log = slog.Default()

func unsetSafe(key string) func() {
	return func() {
		_ = os.Unsetenv(key)
	}
}

func registerCleanupEnv(t *testing.T) {
	t.Cleanup(unsetSafe(parentShared))
	t.Cleanup(unsetSafe(parentOnly))
	t.Cleanup(unsetSafe("DOTENV_FILE"))
	t.Cleanup(unsetSafe("DOTENV_DIR"))
}

func Test__searchAndLoadEnv(t *testing.T) {
	registerCleanupEnv(t)
	cwd, err := os.Getwd()
	require.NoError(t, err)
	testdata := filepath.Join(cwd, "testdata")
	starting := filepath.Join(testdata, "multi", "level", "none")
	err = searchAndLoadEnv(log, starting, lo.ToPtr(3))
	require.ErrorIs(t, err, ErrMaxDepthReached)
	require.Equal(t, "2", os.Getenv(parentShared))
	require.Equal(t, "1", os.Getenv(parentOnly))
}

func Test__searchAndLoadEnvSameLevel(t *testing.T) {
	registerCleanupEnv(t)
	cwd, err := os.Getwd()
	require.NoError(t, err)
	testdata := filepath.Join(cwd, "testdata")
	starting := filepath.Join(testdata, "multi", "level")
	err = searchAndLoadEnv(log, starting, lo.ToPtr(3))
	require.ErrorIs(t, err, ErrMaxDepthReached)
	require.Equal(t, "2", os.Getenv(parentShared))
	require.Equal(t, "1", os.Getenv(parentOnly))
}

func TestAutoLoadEnv__SingleFile(t *testing.T) {
	registerCleanupEnv(t)
	cwd, err := os.Getwd()
	require.NoError(t, err)
	testdata := filepath.Join(cwd, "testdata")
	starting := filepath.Join(testdata, "multi", "level")
	envFile := filepath.Join(starting, ".env")

	err = os.Setenv("DOTENV_FILE", envFile)
	require.NoError(t, err)

	AutoLoadDotEnv(log)
	require.Equal(t, "1", os.Getenv(parentShared))
	require.Equal(t, "", os.Getenv(parentOnly))
}

func TestAutoLoadEnv__Dir(t *testing.T) {
	registerCleanupEnv(t)
	cwd, err := os.Getwd()
	require.NoError(t, err)
	testdata := filepath.Join(cwd, "testdata")
	starting := filepath.Join(testdata, "multi", "level")

	err = os.Setenv("DOTENV_DIR", starting)
	require.NoError(t, err)

	AutoLoadDotEnv(log)
	require.Equal(t, "2", os.Getenv(parentShared))
	require.Equal(t, "1", os.Getenv(parentOnly))
}
