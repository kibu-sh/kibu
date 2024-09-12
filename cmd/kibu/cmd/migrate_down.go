package cmd

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/kibu-sh/kibu/cmd/kibu/cmd/cliflags"
	"github.com/spf13/cobra"
)

type MigrateDownCmd struct {
	*cobra.Command
}

type NewMigrateDownCmdParams struct{}

func NewMigrateDownCmd(params NewMigrateDownCmdParams) (cmd MigrateDownCmd) {
	cmd.Command = &cobra.Command{
		Use:   "down",
		Short: "down",
		Long:  `down`,
		RunE:  newMigrateDownRunE(),
	}
	return
}

func newMigrateDownRunE() RunE {
	return func(cmd *cobra.Command, args []string) (err error) {
		m, err := migrate.New(
			cliflags.MigrateDir.Value(),
			cliflags.MigrateDatabaseUrl.Value(),
		)
		if err != nil {
			return
		}
		return m.Down()
	}
}
