package cmd

import (
	"github.com/discernhq/devx/cmd/devx/cmd/cliflags"
	"github.com/spf13/cobra"
)

type MigrateCmd struct {
	*cobra.Command
}

type NewMigrateCmdParams struct {
	MigrateUpCmd   MigrateUpCmd
	MigrateDownCmd MigrateDownCmd
}

func NewMigrateCmd(params NewMigrateCmdParams) (cmd MigrateCmd) {
	cmd.Command = &cobra.Command{
		Use:   "migrate",
		Short: "migrate",
		Long:  `migrate`,
	}

	_ = cliflags.MigrateDir.BindToCommand(cmd.Command)
	_ = cliflags.MigrateDatabaseUrl.BindToCommand(cmd.Command)

	cmd.AddCommand(params.MigrateUpCmd.Command)
	cmd.AddCommand(params.MigrateDownCmd.Command)

	return
}
