package cmd

import (
	"github.com/kibu-sh/kibu/cmd/kibu/cmd/cliflags"
	"github.com/kibu-sh/kibu/pkg/appcontext"
	"github.com/kibu-sh/kibu/pkg/config"
	"github.com/kibu-sh/kibu/pkg/workspace"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"log/slog"
)

type ConfigCopyCmd struct {
	*cobra.Command
}

type NewConfigCopyCmdParams struct {
	Logger       *slog.Logger
	configLoader configLoaderFunc
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
		cfg, err := params.configLoader()
		if err != nil {
			return
		}

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
