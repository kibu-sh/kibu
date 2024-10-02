package cmd

import (
	"github.com/kibu-sh/kibu/cmd/kibu/cmd/cliflags"
	"github.com/spf13/cobra"
)

func init() {
	cobra.OnInitialize(func() {})
}

type RootCmd struct {
	*cobra.Command
}

type RootCmdParams struct {
	ConfigCmd  ConfigCmd
	BuildCmd   BuildCmd
	MigrateCmd MigrateCmd
	DevCmd     DevCmd
}

func NewRootCmd(params RootCmdParams) (root RootCmd) {
	root.Command = &cobra.Command{
		Use:   "kibu",
		Short: "kibu is a backend development engine for developer productivity",
		Long:  `kibu is a backend development engine for developer productivity`,
	}

	// TODO: don't ignore these
	_ = cliflags.Environment.BindToCommand(root.Command)
	_ = cliflags.Debug.BindToCommand(root.Command)

	root.AddCommand(params.DevCmd.Command)
	root.AddCommand(params.ConfigCmd.Command)
	root.AddCommand(params.MigrateCmd.Command)
	root.AddCommand(params.BuildCmd.Command)

	return
}
