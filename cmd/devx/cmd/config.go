package cmd

import (
	"github.com/discernhq/devx/internal/build"
	"github.com/discernhq/devx/pkg/config"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"os"
)

type NewConfigCmdParams struct {
	fx.In
	RootCmd *cobra.Command `name:"rootCmd"`
}

type NewConfigCmdResult struct {
	fx.Out
	ConfigCmd *cobra.Command `name:"configCmd"`
}

func NewConfigCmd(params NewConfigCmdParams) (result NewConfigCmdResult) {
	result.ConfigCmd = &cobra.Command{
		Use:   "config",
		Short: "config",
		Long:  `config`,
	}
	params.RootCmd.AddCommand(result.ConfigCmd)
	return
}

type NewConfigGetCmdParams struct {
	fx.In
	ConfigStore config.Store
	ConfigCmd   *cobra.Command `name:"configCmd"`
}

type NewConfigGetCmdResult struct {
	fx.Out
	ConfigGetCmd *cobra.Command `name:"configGetCmd"`
}

func NewConfigGetCmd(params NewConfigGetCmdParams) (result NewConfigGetCmdResult) {
	result.ConfigGetCmd = &cobra.Command{
		Use:   "get",
		Short: "get",
		Long:  `get`,
		RunE: newConfigGetRunE(newConfigGetRunEParams{
			ConfigStore: params.ConfigStore,
		}),
	}

	params.ConfigCmd.AddCommand(result.ConfigGetCmd)
	return
}

type newConfigGetRunEParams struct {
	ConfigStore config.Store
}

func newConfigGetRunE(params newConfigGetRunEParams) RunE {
	return func(cmd *cobra.Command, args []string) (err error) {
		cwd, err := os.Getwd()
		if err != nil {
			return
		}

		b, err := build.NewWithDefaults(cwd)
		if err != nil {
			return
		}

		err = b.Exec()

		return
	}
}
