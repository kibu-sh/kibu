package cmd

import (
	"github.com/discernhq/devx/cmd/devx/cmd/cliflags"
	"github.com/discernhq/devx/pkg/appcontext"
	"github.com/discernhq/devx/pkg/config"
	"github.com/discernhq/devx/pkg/workspace"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"log/slog"
)

type ConfigCopyCmd struct {
	*cobra.Command
}

type NewConfigCopyCmdParams struct {
	WorkspaceConfig *workspace.Config
	Logger          *slog.Logger
}

func NewConfigCopyCmd(params NewConfigCopyCmdParams) (cmd ConfigCopyCmd) {
	cmd.Command = &cobra.Command{
		Use:     "copy",
		Short:   "copy",
		Long:    `copy`,
		RunE:    newConfigCopyRunE(params),
		PreRunE: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	}
	_ = cliflags.GoogleProject.BindToCommand(cmd.Command)
	_ = cliflags.ConfigSyncRecursive.BindToCommand(cmd.Command)
	_ = cliflags.ConfigSyncSrcEnv.BindToCommand(cmd.Command)
	_ = cliflags.ConfigSyncDestEnv.BindToCommand(cmd.Command)
	return
}

func newConfigCopyRunE(params NewConfigCopyCmdParams) RunE {
	return func(cmd *cobra.Command, args []string) (err error) {
		sourcePath := args[0]
		ctx := appcontext.Context()
		cfg := params.WorkspaceConfig
		srcEnv := cliflags.ConfigSyncSrcEnv.Value()
		destEnv := cliflags.ConfigSyncDestEnv.Value()
		recursive := cliflags.ConfigSyncRecursive.Value()
		srcStore, err := workspace.NewEnvScopedFileStore(ctx, cfg, srcEnv)
		if err != nil {
			return
		}

		destStore, err := workspace.NewEnvScopedFileStore(ctx, cfg, destEnv)
		if err != nil {
			return
		}

		destKey, err := cfg.ConfigStore.KeyByEnv(destEnv)
		if err != nil {
			return
		}

		copyParams := config.CopyParams{
			Source:         srcStore,
			SourcePath:     sourcePath,
			Destination:    destStore,
			DestinationKey: destKey,
		}

		var copyFn config.CopyFunc

		if recursive {
			slog.Default().Info("copying config recursively")
			copyFn = config.CopyRecursive
		} else {
			copyFn = config.CopyOne
		}

		if err = copyFn(ctx, copyParams); err != nil {
			err = errors.Wrap(err, "failed to copy config")
		}

		return
	}
}
