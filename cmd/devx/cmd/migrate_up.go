package cmd

import (
	"github.com/discernhq/devx/cmd/devx/cmd/cliflags"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/spf13/cobra"
)

type MigrateUpCmd struct {
	*cobra.Command
}

type NewMigrateUpCmdParams struct{}

func NewMigrateUpCmd(params NewMigrateUpCmdParams) (cmd MigrateUpCmd) {
	cmd.Command = &cobra.Command{
		Use:   "up",
		Short: "up",
		Long:  `up`,
		RunE:  newMigrateUpRunE(),
	}
	return
}

func newMigrateUpRunE() RunE {
	return func(cmd *cobra.Command, args []string) (err error) {
		m, err := migrate.New(
			cliflags.MigrateDir.Value(),
			cliflags.MigrateDatabaseUrl.Value(),
		)
		if err != nil {
			return
		}
		return m.Up()
	}
}
