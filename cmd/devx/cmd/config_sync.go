package cmd

import (
	"fmt"
	"github.com/discernhq/devx/cmd/devx/cmd/cliflags"
	"github.com/discernhq/devx/pkg/appcontext"
	"github.com/discernhq/devx/pkg/config"
	"github.com/discernhq/devx/pkg/workspace"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
	"path/filepath"
)

type ConfigSyncCmd struct {
	*cobra.Command
}

type NewConfigSyncCmdParams struct {
	ConfigStoreSettings *workspace.ConfigStoreSettings
	Store               config.Store
}

func NewConfigSyncCmd(params NewConfigSyncCmdParams) (cmd ConfigSyncCmd) {
	cmd.Command = &cobra.Command{
		Use:   "sync",
		Short: "sync",
		Long:  `sync`,
		RunE:  newConfigSyncRunE(params),
	}
	_ = cliflags.GoogleProject.BindToCommand(cmd.Command)
	return
}

func newConfigSyncRunE(params NewConfigSyncCmdParams) RunE {
	return func(cmd *cobra.Command, args []string) (err error) {
		// TODO: this kinda smells
		// its not compatible with any other target stores
		ctx := appcontext.Context()
		store := params.Store.(*config.FileStore)
		storeDir := store.FS.(config.DirectoryFS)
		env := cliflags.Environment.Value()
		envDir := filepath.Join(storeDir.Path, env)
		dirFS := os.DirFS(envDir)
		projectID := cliflags.GoogleProject.Value()

		if projectID == "" {
			err = errors.New("must supply --google-project")
			return
		}

		remoteStore, err := config.NewGCPSecretManagerStore(ctx, projectID)
		if err != nil {
			return
		}

		fmt.Println("attempting to push secrets to google secret manager")

		err = fs.WalkDir(dirFS, ".", func(path string, d fs.DirEntry, _ error) error {
			if d.IsDir() {
				return nil
			}

			fmt.Println(path)

			var data any
			_, readErr := store.Get(ctx, config.GetParams{
				Path: joinSecretEnvPath(joinSecretEnvParams{
					Env:  env,
					Path: path,
				}),
				Result: &data,
			})
			if readErr != nil {
				return readErr
			}

			_, writeErr := remoteStore.Set(ctx, config.SetParams{
				Path: path,
				Data: data,
			})

			if writeErr != nil {
				return writeErr
			}

			return nil
		})

		return
	}
}