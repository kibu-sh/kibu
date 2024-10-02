package testingx

import (
	"context"
	"github.com/google/wire"
	"github.com/kibu-sh/kibu/pkg/config"
	"github.com/kibu-sh/kibu/pkg/workspace"
	"log"
)

func NewWorkspaceFileStore(ctx context.Context) (store *config.FileStore, err error) {
	wsConfig, err := workspace.NewWorkspaceConfig()
	if err != nil {
		return
	}

	store, err = workspace.NewEnvScopedFileStore(ctx, wsConfig, "dev")
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
