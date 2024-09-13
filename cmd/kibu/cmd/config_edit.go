package cmd

import (
	"context"
	"github.com/kibu-sh/kibu/cmd/kibu/cmd/cliflags"
	"github.com/kibu-sh/kibu/pkg/config"
	"github.com/spf13/cobra"
)

type ConfigEditCmd struct {
	*cobra.Command
}

type fileEditorFunc func(config.Store) *config.EncryptedFileEditor

func provideFileEditor() fileEditorFunc {
	return func(store config.Store) *config.EncryptedFileEditor {
		return config.NewEncryptedFileEditor(store)
	}
}

type NewConfigEditCmdParams struct {
	loadStore         storeLoaderFunc
	loadFileEditor    fileEditorFunc
	loadStoreSettings storeSettingsLoaderFunc
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
		store, err := params.loadStore()
		if err != nil {
			return
		}

		settings, err := params.loadStoreSettings()
		if err != nil {
			return
		}

		editor := params.loadFileEditor(store)

		key, err := settings.KeyByEnv(cliflags.Environment.Value())
		if err != nil {
			return err
		}

		path := joinSecretEnvPath(joinSecretEnvParams{
			Env:  cliflags.Environment.Value(),
			Path: args[0],
		})

		err = editor.Edit(context.Background(), config.EditParams{
			Path:          path,
			EncryptionKey: key,
		})

		return
	}
}
