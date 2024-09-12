package cmd

import (
	"github.com/spf13/cobra"
)

type RunE func(cmd *cobra.Command, args []string) error
