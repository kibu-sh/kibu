package cmd

import (
	"context"
	"encoding/json"
	"github.com/kibu-sh/kibu/cmd/kibu/cmd/cliflags"
	"github.com/kibu-sh/kibu/pkg/config"
	"github.com/kibu-sh/kibu/pkg/workspace"
	"github.com/spf13/cobra"
)

type ConfigGetCmd struct {
	*cobra.Command
}

type storeLoaderFunc func() (config.Store, error)
type configLoaderFunc func() (*workspace.Config, error)

type NewConfigGetCmdParams struct {
	StoreLoader storeLoaderFunc
}

func provideConfigLoader() configLoaderFunc {
	return func() (*workspace.Config, error) {
		return workspace.NewWorkspaceConfig()
	}
}

func provideStoreLoader(ctx context.Context, configLoader configLoaderFunc) storeLoaderFunc {
	return func() (config.Store, error) {
		workspaceConfig, err := configLoader()
		if err != nil {
			return nil, err
		}
		return workspace.NewFileStore(ctx, workspaceConfig)
	}
}

func NewConfigGetCmd(params NewConfigGetCmdParams) (cmd ConfigGetCmd) {
	cmd.Command = &cobra.Command{
		Use:     "get",
		Short:   "get",
		Long:    `get`,
		PreRunE: cobra.ExactValidArgs(1),
		RunE:    newConfigGetRunE(params),
	}
	return
}

func newConfigGetRunE(params NewConfigGetCmdParams) RunE {
	return func(cmd *cobra.Command, args []string) (err error) {
		var data any
		path := joinSecretEnvPath(joinSecretEnvParams{
			Env:  cliflags.Environment.Value(),
			Path: args[0],
		})

		store, err := params.StoreLoader()
		if err != nil {
			return
		}

		_, err = store.Get(context.Background(), config.GetParams{
			Path:   path,
			Result: &data,
		})

		if err != nil {
			return
		}

		out, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return
		}

		_, err = cmd.OutOrStdout().Write(out)

		return
	}
}
