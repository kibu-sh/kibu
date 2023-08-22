package cmd

import (
	"github.com/discernhq/devx/pkg/appcontext"
	"github.com/discernhq/devx/pkg/config"
	"github.com/discernhq/devx/pkg/workspace"
	"github.com/google/wire"
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
