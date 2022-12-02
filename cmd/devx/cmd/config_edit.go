package cmd

import (
	"context"
	"github.com/discernhq/devx/cmd/devx/cmd/cliflags"
	"github.com/discernhq/devx/internal/workspace"
	"github.com/discernhq/devx/pkg/config"
	"github.com/spf13/cobra"
)

type ConfigEditCmd struct {
	*cobra.Command
}

type NewConfigEditCmdParams struct {
	ConfigStoreSettings *workspace.ConfigStoreSettings
	EncryptedFileEditor *config.EncryptedFileEditor
}

func NewConfigEditCmd(params NewConfigEditCmdParams) (cmd ConfigEditCmd) {
	cmd.Command = &cobra.Command{
		Use:     "edit",
		Short:   "edit",
		Long:    `edit`,
		PreRunE: cobra.ExactValidArgs(1),
		RunE:    newConfigEditRunE(params),
	}
	return
}

func newConfigEditRunE(params NewConfigEditCmdParams) RunE {
	return func(cmd *cobra.Command, args []string) (err error) {
		key, err := params.ConfigStoreSettings.KeyByEnv(cliflags.Environment.Value())
		if err != nil {
			return err
		}

		err = params.EncryptedFileEditor.Edit(context.Background(), config.EditParams{
			Path:          args[0],
			EncryptionKey: key,
		})

		return
	}
}
