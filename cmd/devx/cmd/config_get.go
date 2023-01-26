package cmd

import (
	"context"
	"encoding/json"
	"github.com/discernhq/devx/cmd/devx/cmd/cliflags"
	"github.com/discernhq/devx/pkg/config"
	"github.com/spf13/cobra"
)

type ConfigGetCmd struct {
	*cobra.Command
}

type NewConfigGetCmdParams struct {
	Store config.Store
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

		_, err = params.Store.Get(context.Background(), config.GetParams{
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
