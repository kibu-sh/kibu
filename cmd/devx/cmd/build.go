package cmd

import (
	"github.com/discernhq/devx/internal/codegen"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
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

		err = codegen.Generate(codegen.GenerateParams{
			Dir:       cwd,
			Patterns:  args,
			Pipeline:  codegen.DefaultPipeline(),
			OutputDir: filepath.Join(cwd, "gen/devxgen"),
		})
		return
	}
}
