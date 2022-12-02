package cmd

import (
	"context"
	"github.com/discernhq/devx/internal/workspace"
	"github.com/discernhq/devx/pkg/config"
	"github.com/spf13/cobra"
	"path/filepath"
)

func NewWorkspaceConfig() (*workspace.Config, error) {
	return workspace.LoadConfigFromCWD(workspace.LoadConfigParams{
		DetermineRootParams: workspace.DetermineRootParams{
			SearchSuffix: ".devx/workspace.cue",
		},
		LoaderFunc: workspace.CueLoader,
	})
}

func NewConfigFileStore(ctx context.Context, ws *workspace.Config) (*config.FileStore, error) {
	return config.NewDefaultFileStore(filepath.Join(ws.ConfigRoot(), "store/config")), nil
}

type RunE func(cmd *cobra.Command, args []string) error
