package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func init() {
	cobra.OnInitialize(func() {

	})
}

type NewRootCmdResult struct {
	fx.Out
	RootCmd *cobra.Command `name:"rootCmd"`
}

func NewRootCmd() (result NewRootCmdResult) {
	result.RootCmd = &cobra.Command{
		Use:   "devx",
		Short: "devx is a backend development engine for developer productivity",
		Long:  `devx is a backend development engine for developer productivity`,
	}
	return
}

var Module = fx.Module("cmd",
	fx.Provide(NewRootCmd),
	fx.Provide(NewBuildCmd),
	fx.Provide(NewConfigCmd),
	fx.Provide(NewConfigGetCmd),
)
