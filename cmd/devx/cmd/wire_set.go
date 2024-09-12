package cmd

import (
	"github.com/google/wire"
	"github.com/kibu-sh/kibu/pkg/appcontext"
	"github.com/kibu-sh/kibu/pkg/config"
	"github.com/kibu-sh/kibu/pkg/workspace"
	"log/slog"
	"os"
)

func NewLogger() (l *slog.Logger) {
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}

var WireSet = wire.NewSet(
	appcontext.Context,
	workspace.NewWorkspaceConfig,
	workspace.NewFileStore,
	config.NewEncryptedFileEditor,

	NewLogger,

	NewRootCmd,
	NewBuildCmd,
	NewConfigCmd,
	NewConfigGetCmd,
	NewConfigSetCmd,
	NewConfigEditCmd,
	NewConfigCopyCmd,
	NewConfigSyncCmd,
	NewMigrateCmd,
	NewMigrateUpCmd,
	NewMigrateDownCmd,

	wire.Bind(new(config.Store), new(*config.FileStore)),
	wire.Struct(new(RootCommandParams), "*"),
	wire.Struct(new(ConfigCmdParams), "*"),
	wire.Struct(new(NewConfigGetCmdParams), "*"),
	wire.Struct(new(NewConfigSetCmdParams), "*"),
	wire.Struct(new(NewConfigEditCmdParams), "*"),
	wire.Struct(new(NewConfigSyncCmdParams), "*"),
	wire.Struct(new(NewMigrateCmdParams), "*"),
	wire.Struct(new(NewMigrateDownCmdParams), "*"),
	wire.Struct(new(NewMigrateUpCmdParams), "*"),
	wire.Struct(new(NewConfigCopyCmdParams), "*"),
	wire.FieldsOf(new(*workspace.Config), "ConfigStore"),
)
