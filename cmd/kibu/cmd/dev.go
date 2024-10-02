package cmd

import (
	"github.com/spf13/cobra"
)

type DevCmd struct {
	*cobra.Command
}

type DevCmdParams struct {
	DevUpCmd DevUpCmd
}

func NewDevCmd(params DevCmdParams) (cmd DevCmd) {
	cmd.Command = &cobra.Command{
		Use:   "dev",
		Short: "dev",
		Long:  `dev`,
	}

	cmd.AddCommand(params.DevUpCmd.Command)
	return
}
