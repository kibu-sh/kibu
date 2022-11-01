package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "devx",
		Short: "devx is a backend development engine for developer productivity",
		Long:  `devx is a backend development engine for developer productivity`,
	}
}

func Init() (root *cobra.Command, err error) {
	root = NewRootCmd()
	root.AddCommand(NewBuildCmd())
	return
}

func init() {
	cobra.OnInitialize(func() {

	})
}
