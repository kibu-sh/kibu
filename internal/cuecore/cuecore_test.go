package cuecore

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	testdata := filepath.Join(cwd, "../testdata")
	workflowFile := filepath.Join(testdata, "module.cue")

	_, err = Load(LoadOptions{
		Dir:        cwd,
		Entrypoint: []string{workflowFile},
	})
	require.NoError(t, err)
}
