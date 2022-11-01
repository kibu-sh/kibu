package cuemod

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {
	dir, err := os.MkdirTemp("", "devx")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	err = Copy(filepath.Join(dir, "cue.mod"))
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(dir, "cue.mod/module.cue"))
	require.FileExists(t, filepath.Join(dir, "cue.mod/pkg/discern.com/devx/devx.cue"))
}
