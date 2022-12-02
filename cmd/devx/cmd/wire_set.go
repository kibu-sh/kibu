package cmd

import (
	"github.com/discernhq/devx/internal/workspace"
	"github.com/discernhq/devx/pkg/appcontext"
	"github.com/discernhq/devx/pkg/config"
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	appcontext.Context,
	NewWorkspaceConfig,
	NewRootCmd,
	NewBuildCmd,
	NewConfigCmd,
	NewConfigGetCmd,
	NewConfigSetCmd,
	NewConfigFileStore,
	NewConfigEditCmd,
	config.NewEncryptedFileEditor,

	wire.Bind(new(config.Store), new(*config.FileStore)),
	wire.Struct(new(RootCommandParams), "*"),
	wire.Struct(new(ConfigCmdParams), "*"),
	wire.Struct(new(NewConfigGetCmdParams), "*"),
	wire.Struct(new(NewConfigSetCmdParams), "*"),
	wire.Struct(new(NewConfigEditCmdParams), "*"),
	wire.FieldsOf(new(*workspace.Config), "ConfigStore"),
)
