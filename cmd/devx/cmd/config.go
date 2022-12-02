package cmd

import (
	"github.com/spf13/cobra"
)

type ConfigCmd struct {
	*cobra.Command
}

type ConfigCmdParams struct {
	ConfigGetCmd  ConfigGetCmd
	ConfigSetCmd  ConfigSetCmd
	ConfigEditCmd ConfigEditCmd
}

func NewConfigCmd(params ConfigCmdParams) (cmd ConfigCmd) {
	cmd.Command = &cobra.Command{
		Use:   "config",
		Short: "config",
		Long:  `config`,
	}

	cmd.AddCommand(params.ConfigGetCmd.Command)
	cmd.AddCommand(params.ConfigSetCmd.Command)
	cmd.AddCommand(params.ConfigEditCmd.Command)
	return
}
