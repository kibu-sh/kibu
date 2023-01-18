package testutilsx

import (
	"context"
	"github.com/discernhq/devx/pkg/config"
	"github.com/discernhq/devx/pkg/workspace"
	"github.com/google/wire"
	"log"
)

func NewWorkspaceFileStore(ctx context.Context) (store *config.FileStore, err error) {
	wsConfig, err := workspace.NewWorkspaceConfig()
	if err != nil {
		return
	}

	store, err = workspace.NewFileStore(ctx, wsConfig)
	return
}

func CheckErrFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var LocalFileStoreSet = wire.NewSet(
	NewWorkspaceFileStore,
	wire.Bind(new(config.Store), new(*config.FileStore)),
)
