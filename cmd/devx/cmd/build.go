package cmd

import (
	"github.com/discernhq/devx/internal/build"
	"github.com/spf13/cobra"
	"os"
)

type BuildCmd struct {
	*cobra.Command
}

func NewBuildCmd() (cmd BuildCmd) {
	cmd.Command = &cobra.Command{
		Use:   "build",
		Short: "build code",
		Long:  `build code`,
		RunE:  newBuildRunE(),
	}
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
