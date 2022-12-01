package cmd

import (
	"github.com/discernhq/devx/internal/build"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"os"
)

type RunE func(cmd *cobra.Command, args []string) error

type NewBuildCmdParams struct {
	fx.In
	RootCmd *cobra.Command `name:"rootCmd"`
}

type NewBuildCmdResult struct {
	fx.Out
	BuildCmd *cobra.Command `name:"buildCmd"`
}

func NewBuildCmd(params NewBuildCmdParams) (result NewBuildCmdResult) {
	result.BuildCmd = &cobra.Command{
		Use:   "build",
		Short: "build code",
		Long:  `build code`,
		RunE:  newBuildRunE(),
	}

	params.RootCmd.AddCommand(result.BuildCmd)
	return
}

func newBuildRunE() RunE {
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
