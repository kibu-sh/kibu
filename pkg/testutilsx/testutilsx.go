package testutilsx

import (
	"context"
	"github.com/discernhq/devx/pkg/config"
	"github.com/discernhq/devx/pkg/workspace"
)

func NewWorkspaceFileStore(ctx context.Context) (store *config.FileStore, err error) {
	wsConfig, err := workspace.NewWorkspaceConfig()
	if err != nil {
		return
	}

	store, err = workspace.NewFileStore(ctx, wsConfig)
	return
}
